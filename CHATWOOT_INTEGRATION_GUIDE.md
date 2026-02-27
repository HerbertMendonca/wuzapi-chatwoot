# Guia de Integração Wuzapi com Chatwoot

Este guia detalha os passos necessários para configurar a integração entre o Wuzapi (com as modificações implementadas) e o Chatwoot.

## 1. Configuração do Chatwoot

Para que o Chatwoot possa se comunicar com o Wuzapi, você precisará configurar um webhook de saída e obter as credenciais necessárias.

### 1.1 Criar um Canal WhatsApp no Chatwoot

1.  No seu painel do Chatwoot, navegue até **Settings (Configurações)** > **Inboxes (Caixas de Entrada)**.
2.  Clique em **Add Inbox (Adicionar Caixa de Entrada)**.
3.  Selecione **WhatsApp** como o tipo de canal.
4.  Escolha **API (On-premise)** como provedor.
5.  Preencha os detalhes do seu canal WhatsApp, como nome e agentes atribuídos.
6.  Após a criação, você será direcionado para a página de configurações do Inbox. Anote o **Inbox ID** e o **Account ID** (disponíveis na URL ou nas informações do Inbox).

### 1.2 Obter o Token de Acesso do Chatwoot

1.  No Chatwoot, navegue até **Settings (Configurações)** > **Profile (Perfil)**.
2.  Role para baixo até a seção **API Access Tokens (Tokens de Acesso à API)**.
3.  Gere um novo token de acesso ou utilize um existente. Anote este **Access Token**, pois ele será usado pelo Wuzapi para se autenticar no Chatwoot.

### 1.3 Configurar o Webhook de Saída no Chatwoot

1.  No Chatwoot, navegue até **Settings (Configurações)** > **Inboxes (Caixas de Entrada)** e selecione o Inbox do WhatsApp que você criou.
2.  Vá para a aba **Webhooks**.
3.  Clique em **Add Webhook (Adicionar Webhook)**.
4.  No campo **Webhook URL**, insira a URL do seu Wuzapi para o endpoint de webhook do Chatwoot. O formato será algo como:
    `https://wuzapi.zorbix.cloud/chatwoot/webhook?token=SEU_WUZAPI_TOKEN`
    *   Substitua `SEU_IP_OU_DOMINIO_WUZAPI` pelo endereço IP ou domínio onde seu Wuzapi está rodando.
    *   Substitua `PORTA` pela porta que o Wuzapi está escutando (padrão 8080).
    *   Substitua `SEU_WUZAPI_TOKEN` pelo token de usuário do Wuzapi que você deseja associar a este Inbox do Chatwoot. Este token é crucial para o Wuzapi identificar qual instância do WhatsApp deve ser usada para enviar a mensagem.
5.  Selecione os eventos que você deseja que o Chatwoot envie para o Wuzapi. Para a sincronização de mensagens de saída, o evento principal é `message_created`.
6.  Salve o webhook.

## 2. Configuração do Wuzapi

As credenciais do Chatwoot podem ser configuradas no Wuzapi de duas formas: globalmente via variáveis de ambiente no `docker-compose.yml` ou por usuário via API.

### 2.1 Configuração Global (docker-compose.yml)

Você pode definir as variáveis de ambiente no seu arquivo `docker-compose.yml` para que todas as instâncias do Wuzapi usem as mesmas credenciais do Chatwoot. Isso é útil para ambientes com um único Chatwoot ou para definir padrões.

Adicione as seguintes variáveis de ambiente ao serviço `wuzapi-server` no seu `docker-compose.yml`:

```yaml
    environment:
      # ... outras variáveis de ambiente existentes ...
      - CHATWOOT_URL=https://seu.chatwoot.com
      - CHATWOOT_ACCOUNT_ID=SEU_ACCOUNT_ID_CHATWOOT
      - CHATWOOT_ACCESS_TOKEN=SEU_ACCESS_TOKEN_CHATWOOT
      - CHATWOOT_INBOX_ID=SEU_INBOX_ID_CHATWOOT
```

*   Substitua `https://seu.chatwoot.com` pela URL da sua instância do Chatwoot.
*   Substitua `SEU_ACCOUNT_ID_CHATWOOT` pelo Account ID que você anotou no passo 1.1.
*   Substitua `SEU_ACCESS_TOKEN_CHATWOOT` pelo Access Token que você anotou no passo 1.2.
*   Substitua `SEU_INBOX_ID_CHATWOOT` pelo Inbox ID que você anotou no passo 1.1.

### 2.2 Configuração por Usuário (API)

Para maior flexibilidade, você pode configurar as credenciais do Chatwoot para cada usuário do Wuzapi individualmente através de uma nova API. Esta configuração sobrescreverá as variáveis de ambiente globais.

**Endpoint:** `POST /chatwoot/configure`

**Headers:**
*   `Authorization: Bearer SEU_WUZAPI_ADMIN_TOKEN` (ou o token do usuário específico)

**Body (JSON):**

```json
{
  "chatwoot_url": "https://seu.chatwoot.com",
  "chatwoot_account_id": "SEU_ACCOUNT_ID_CHATWOOT",
  "chatwoot_access_token": "SEU_ACCESS_TOKEN_CHATWOOT",
  "chatwoot_inbox_id": "SEU_INBOX_ID_CHATWOOT"
}
```

*   Substitua os valores pelos dados correspondentes do seu Chatwoot.
*   O `SEU_WUZAPI_ADMIN_TOKEN` é o token de administrador do Wuzapi, ou o token de um usuário específico se você quiser configurar apenas para ele.

## 3. Implantação no Portainer

Com o `Dockerfile` e `docker-compose.yml` atualizados, e as configurações do Chatwoot em mãos, você pode implantar o Wuzapi no Portainer.

1.  **Atualize seu repositório Git:** Certifique-se de que os arquivos `Dockerfile`, `docker-compose.yml`, `chatwoot.go`, `wmiau.go`, `handlers.go`, `routes.go` e `migrations.go` (e quaisquer outros arquivos modificados) estejam atualizados no seu repositório Git.
2.  **No Portainer:**
    *   Navegue até **Stacks**.
    *   Clique em **Add stack** ou edite uma stack existente.
    *   Se você estiver criando uma nova stack, selecione **Git Repository** e aponte para o seu repositório onde o Wuzapi modificado está.
    *   Se você estiver atualizando uma stack existente, clique em **Editor** e cole o conteúdo atualizado do `docker-compose.yml` ou use a opção de `Pull and redeploy` se o Portainer estiver configurado para isso.
    *   Certifique-se de definir as variáveis de ambiente necessárias (como `WUZAPI_ADMIN_TOKEN`, `WUZAPI_GLOBAL_ENCRYPTION_KEY`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, etc., e as novas variáveis `CHATWOOT_URL`, `CHATWOOT_ACCOUNT_ID`, `CHATWOOT_ACCESS_TOKEN`, `CHATWOOT_INBOX_ID` se estiver usando a configuração global) no Portainer.
    *   Implante ou atualize a stack.

Após a implantação, o Wuzapi estará rodando com a integração do Chatwoot. Mensagens recebidas no WhatsApp serão enviadas para o Chatwoot, e respostas de agentes no Chatwoot serão enviadas de volta para o WhatsApp.
