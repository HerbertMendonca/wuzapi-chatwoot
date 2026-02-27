package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

type ChatwootConfig struct {
	URL       string `json:"chatwoot_url"`
	AccountID string `json:"chatwoot_account_id"`
	Token     string `json:"chatwoot_token"`
	InboxID   string `json:"chatwoot_inbox_id"`
}

// ChatwootWebhook represents the payload sent by Chatwoot
type ChatwootWebhook struct {
	Event         string `json:"event"`
	ID            int    `json:"id"`
	Content       string `json:"content"`
	MessageType   string `json:"message_type"`
	ContentType   string `json:"content_type"`
	Conversation  struct {
		ID              int `json:"id"`
		ContactInbox    struct {
			SourceID string `json:"source_id"`
		} `json:"contact_inbox"`
	} `json:"conversation"`
	Sender struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"sender"`
	Account struct {
		ID int `json:"id"`
	} `json:"account"`
}

func (s *server) sendToChatwoot(userID string, contactJID string, contactName string, message string, isOutgoing bool) {
	// Get user config from DB or cache
	var config ChatwootConfig
	err := s.db.Get(&config, "SELECT chatwoot_url, chatwoot_account_id, chatwoot_token, chatwoot_inbox_id FROM users WHERE id=$1", userID)
	if err != nil {
		log.Error().Err(err).Str("userID", userID).Msg("Failed to get Chatwoot config from DB")
		return
	}

	if config.URL == "" || config.AccountID == "" || config.Token == "" || config.InboxID == "" {
		return
	}

	// 1. Find or Create Contact
	sourceID := strings.Split(contactJID, "@")[0]
	contactID, err := s.chatwootFindOrCreateContact(config, sourceID, contactName)
	if err != nil {
		log.Error().Err(err).Msg("Chatwoot: Failed to find/create contact")
		return
	}

	// 2. Find or Create Conversation
	convID, err := s.chatwootFindOrCreateConversation(config, contactID)
	if err != nil {
		log.Error().Err(err).Msg("Chatwoot: Failed to find/create conversation")
		return
	}

	// 3. Send Message
	err = s.chatwootSendMessage(config, convID, message, isOutgoing)
	if err != nil {
		log.Error().Err(err).Msg("Chatwoot: Failed to send message")
	}
}

func (s *server) chatwootFindOrCreateContact(config ChatwootConfig, sourceID string, name string) (int, error) {
	url := fmt.Sprintf("%s/api/v1/accounts/%s/contacts/search?q=%s", config.URL, config.AccountID, sourceID)
	
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("api_access_token", config.Token)
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var searchResult struct {
		Payload []struct {
			ID int `json:"id"`
		} `json:"payload"`
	}
	json.NewDecoder(resp.Body).Decode(&searchResult)

	if len(searchResult.Payload) > 0 {
		return searchResult.Payload[0].ID, nil
	}

	// Create contact
	url = fmt.Sprintf("%s/api/v1/accounts/%s/contacts", config.URL, config.AccountID)
	body, _ := json.Marshal(map[string]interface{}{
		"inbox_id": config.InboxID,
		"name":     name,
		"identifier": sourceID,
		"custom_attributes": map[string]string{
			"source_id": sourceID,
		},
	})

	req, _ = http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("api_access_token", config.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var createResult struct {
		Payload struct {
			Contact struct {
				ID int `json:"id"`
			} `json:"contact"`
		} `json:"payload"`
	}
	json.NewDecoder(resp.Body).Decode(&createResult)
	return createResult.Payload.Contact.ID, nil
}

func (s *server) chatwootFindOrCreateConversation(config ChatwootConfig, contactID int) (int, error) {
	// Try to find open conversation
	url := fmt.Sprintf("%s/api/v1/accounts/%s/contacts/%d/conversations", config.URL, config.AccountID, contactID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("api_access_token", config.Token)
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var convs struct {
		Payload []struct {
			ID     int    `json:"id"`
			Status string `json:"status"`
		} `json:"payload"`
	}
	json.NewDecoder(resp.Body).Decode(&convs)

	for _, c := range convs.Payload {
		if c.Status == "open" || c.Status == "pending" {
			return c.ID, nil
		}
	}

	// Create new conversation
	url = fmt.Sprintf("%s/api/v1/accounts/%s/conversations", config.URL, config.AccountID)
	body, _ := json.Marshal(map[string]interface{}{
		"source_id":  fmt.Sprintf("conv_%d", time.Now().Unix()),
		"contact_id": contactID,
		"inbox_id":   config.InboxID,
	})

	req, _ = http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("api_access_token", config.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var createResult struct {
		ID int `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&createResult)
	return createResult.ID, nil
}

func (s *server) chatwootSendMessage(config ChatwootConfig, convID int, message string, isOutgoing bool) error {
	url := fmt.Sprintf("%s/api/v1/accounts/%s/conversations/%d/messages", config.URL, config.AccountID, convID)
	
	messageType := "incoming"
	if isOutgoing {
		messageType = "outgoing"
	}

	body, _ := json.Marshal(map[string]interface{}{
		"content":      message,
		"message_type": messageType,
		"private":      false,
	})

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("api_access_token", config.Token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chatwoot api error: %s", string(b))
	}

	return nil
}

func (s *server) ChatwootWebhook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var webhook ChatwootWebhook
		if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
			s.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		// Only process outgoing messages from agents
		if webhook.Event == "message_created" && webhook.MessageType == "outgoing" && !strings.Contains(webhook.Sender.Type, "contact") {
			// Find which user this belongs to based on Chatwoot Account ID and Inbox ID
			var userID string
			var token string
			// Chatwoot sends inbox_id in the conversation object, but we also have it in the root for some events
			inboxID := webhook.Conversation.ID // This is conversation ID, we need inbox ID
			// Actually, the webhook payload usually has inbox { id: ... } or we can get it from the conversation
			// Let's use a more robust way or rely on the token in the URL
			
			err := s.db.QueryRow("SELECT id, token FROM users WHERE chatwoot_account_id=$1 AND chatwoot_token IS NOT NULL LIMIT 1", 
				fmt.Sprintf("%d", webhook.Account.ID)).Scan(&userID, &token)
			
			// If not found by account/inbox, we might need a better way or a specific token in the URL
			// For now, let's assume the URL has the wuzapi token: /chatwoot/webhook?token=...
			if err != nil {
				token = r.URL.Query().Get("token")
				if token == "" {
					log.Error().Msg("Chatwoot Webhook: Could not identify user")
					w.WriteHeader(http.StatusOK)
					return
				}
				err = s.db.QueryRow("SELECT id FROM users WHERE token=$1 LIMIT 1", token).Scan(&userID)
				if err != nil {
					log.Error().Msg("Chatwoot Webhook: Invalid token")
					w.WriteHeader(http.StatusOK)
					return
				}
			}

			// Send message via WhatsApp
			client := clientManager.GetWhatsmeowClient(userID)
			if client == nil || !client.IsConnected() {
				log.Error().Str("userID", userID).Msg("Chatwoot Webhook: WhatsApp client not connected")
				w.WriteHeader(http.StatusOK)
				return
			}

			targetJID := webhook.Conversation.ContactInbox.SourceID
			if !strings.Contains(targetJID, "@") {
				targetJID = targetJID + "@s.whatsapp.net"
			}

			// Send message via WhatsApp
			_, _, err = s.internalSendMessage(context.Background(), userID, textMessageRequest{
				Phone: targetJID,
				Body:  webhook.Content,
			})
			if err != nil {
				log.Error().Err(err).Msg("Chatwoot Webhook: Failed to send WhatsApp message")
			} else {
				log.Info().Str("to", targetJID).Msg("Message sent from Chatwoot to WhatsApp")
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (s *server) ConfigureChatwoot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		txtid := r.Context().Value("userinfo").(Values).Get("Id")
		var config ChatwootConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			s.Respond(w, r, http.StatusBadRequest, err)
			return
		}

		query := "UPDATE users SET chatwoot_url=$1, chatwoot_account_id=$2, chatwoot_token=$3, chatwoot_inbox_id=$4 WHERE id=$5"
		if s.db.DriverName() == "sqlite" {
			query = "UPDATE users SET chatwoot_url=?, chatwoot_account_id=?, chatwoot_token=?, chatwoot_inbox_id=? WHERE id=?"
		}

		_, err := s.db.Exec(query, config.URL, config.AccountID, config.Token, config.InboxID, txtid)
		if err != nil {
			s.Respond(w, r, http.StatusInternalServerError, err)
			return
		}

		s.Respond(w, r, http.StatusOK, map[string]string{"status": "success", "message": "Chatwoot configuration updated"})
	}
}

func (s *server) GetChatwootConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		txtid := r.Context().Value("userinfo").(Values).Get("Id")
		var config ChatwootConfig
		err := s.db.Get(&config, "SELECT chatwoot_url, chatwoot_account_id, chatwoot_token, chatwoot_inbox_id FROM users WHERE id=$1", txtid)
		if err != nil {
			s.Respond(w, r, http.StatusInternalServerError, err)
			return
		}
		s.Respond(w, r, http.StatusOK, config)
	}
}

func (s *server) DeleteChatwootConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		txtid := r.Context().Value("userinfo").(Values).Get("Id")
		query := "UPDATE users SET chatwoot_url='', chatwoot_account_id='', chatwoot_token='', chatwoot_inbox_id='' WHERE id=$1"
		if s.db.DriverName() == "sqlite" {
			query = "UPDATE users SET chatwoot_url='', chatwoot_account_id='', chatwoot_token='', chatwoot_inbox_id='' WHERE id=?"
		}
		_, err := s.db.Exec(query, txtid)
		if err != nil {
			s.Respond(w, r, http.StatusInternalServerError, err)
			return
		}
		s.Respond(w, r, http.StatusOK, map[string]string{"status": "success", "message": "Chatwoot configuration deleted"})
	}
}
