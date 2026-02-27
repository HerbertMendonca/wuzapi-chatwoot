# Guia de Instalação do Wuzapi com Integração Chatwoot no Portainer

Este guia detalha como implantar a versão modificada do Wuzapi, que inclui a integração com o Chatwoot, no seu ambiente Portainer. Utilizaremos o repositório GitHub que foi criado (`https://github.com/HerbertMendonca/wuzapi-chatwoot`) e as configurações de `docker-compose.yml` adaptadas à sua infraestrutura existente (volumes externos, rede `Nex1Net` e labels do Traefik).

## Pré-requisitos

*   **Repositório GitHub Atualizado:** Certifique-se de que o repositório `https://github.com/HerbertMendonca/wuzapi-chatwoot` contém os arquivos `Dockerfile` e `docker-compose.yml` atualizados, bem como os arquivos de código-fonte modificados (`chatwoot.go`, `handlers.go`, `routes.go`, `wmiau.go`, `migrations.go`).
*   **Portainer:** Acesso ao seu painel do Portainer.
*   **Volumes e Rede Existentes:** Os volumes `wuzapi_dbdata` e `wuzapi_files`, e a rede `Nex1Net` devem existir no seu ambiente Docker, conforme sua configuração atual.
*   **Servidor PostgreSQL:** Você já possui um serviço PostgreSQL (`db`) configurado e acessível na rede `Nex1Net`, conforme indicado no seu `docker-compose.yml` original.
*   **RabbitMQ:** Você já possui um serviço RabbitMQ configurado e acessível na rede `Nex1Net`, conforme indicado no seu `docker-compose.yml` original.

## Passos para Instalação/Atualização no Portainer

### 1. Acessar o Portainer

Faça login no seu painel do Portainer.

### 2. Criar ou Atualizar uma Stack

No Portainer, a maneira mais eficiente de gerenciar aplicações Docker Compose é através de Stacks.

#### Opção A: Criar uma Nova Stack (Recomendado para primeira instalação ou migração limpa)

1.  No menu lateral, navegue até **Stacks**.
2.  Clique em **Add stack**.
3.  **Name (Nome):** Dê um nome para sua stack, por exemplo, `wuzapi-chatwoot`.
4.  **Git Repository:**
    *   Selecione a opção **Git Repository**.
    *   **Repository URL:** `https://github.com/HerbertMendonca/wuzapi-chatwoot.git`
    *   **Reference (branch/tag):** `main` (ou a branch que você estiver usando).
    *   **Authentication:** Se o seu repositório for privado, você precisará configurar a autenticação (geralmente um Personal Access Token do GitHub).
    *   **Compose path:** `docker-compose.yml` (este é o caminho padrão, mas verifique se o arquivo está na raiz do seu repositório).
5.  **Environment variables (Variáveis de Ambiente):**
    *   Você precisará adicionar as variáveis de ambiente que o Wuzapi utiliza, incluindo as credenciais do Chatwoot. É **altamente recomendável** que você defina essas variáveis diretamente no Portainer para evitar expor informações sensíveis no seu `docker-compose.yml` no GitHub.
    *   Clique em **Add environment variable** e adicione as seguintes (substitua os valores pelos seus):
        *   `WUZAPI_ADMIN_TOKEN`: Seu token de administrador do Wuzapi.
        *   `SECRET_KEY`: Sua chave secreta para criptografia.
        *   `WUZAPI_GLOBAL_ENCRYPTION_KEY`: Chave de criptografia global (se aplicável).
        *   `WUZAPI_GLOBAL_HMAC_KEY`: Chave HMAC global (se aplicável).
        *   `WUZAPI_GLOBAL_WEBHOOK`: Webhook global (se aplicável).
        *   `DB_HOST`: `postgres` (ou o nome do serviço do seu banco de dados PostgreSQL).
        *   `DB_USER`: `postgres` (ou seu usuário do banco de dados).
        *   `DB_PASSWORD`: Sua senha do banco de dados.
        *   `DB_NAME`: `wuzapi` (ou o nome do seu banco de dados).
        *   `DB_PORT`: `5432`.
        *   `DB_DRIVER`: `postgres`.
        *   `TZ`: `America/Sao_Paulo`.
        *   `WEBHOOK_FORMAT`: `json`.
        *   `SESSION_DEVICE_NAME`: `WuzAPI`.
        *   `CHATWOOT_URL`: `https://seu.chatwoot.com` (URL da sua instância do Chatwoot).
        *   `CHATWOOT_ACCOUNT_ID`: O ID da sua conta no Chatwoot.
        *   `CHATWOOT_ACCESS_TOKEN`: O token de acesso à API do Chatwoot.
        *   `CHATWOOT_INBOX_ID`: O ID da caixa de entrada do WhatsApp no Chatwoot.
        *   `RABBITMQ_URL`: `amqp://wuzapi:wuzapi@rabbitmq:5672/` (se estiver usando RabbitMQ).
        *   `RABBITMQ_QUEUE`: `whatsapp_events` (se estiver usando RabbitMQ).
6.  Clique em **Deploy the stack**.

#### Opção B: Atualizar uma Stack Existente (Se você já tem o Wuzapi rodando como uma Stack no Portainer)

1.  No menu lateral, navegue até **Stacks**.
2.  Clique na stack existente do seu Wuzapi.
3.  Clique em **Editor**.
4.  Substitua o conteúdo do editor pelo `docker-compose.yml` adaptado que você tem no seu repositório GitHub. Certifique-se de que as variáveis de ambiente estejam configuradas corretamente (seja no arquivo ou nas variáveis de ambiente da stack).
5.  Alternativamente, se sua stack foi criada a partir de um repositório Git, você pode usar a opção **Pull and redeploy** para puxar as últimas alterações do seu repositório e reconstruir o serviço.
6.  Clique em **Update the stack**.

### 3. Verificar o Status da Implantação

Após a implantação, o Portainer irá construir a imagem Docker a partir do seu `Dockerfile` e iniciar os serviços. Você pode monitorar o progresso na seção **Logs** dos seus contêineres.

Certifique-se de que o contêiner `wuzapi` esteja `running` e sem erros nos logs.

## Próximos Passos

Após a implantação bem-sucedida do Wuzapi no Portainer, você precisará configurar o Chatwoot para interagir com o Wuzapi. Consulte o guia `CHATWOOT_INTEGRATION_GUIDE.md` (disponível no seu repositório GitHub) para os detalhes de configuração do Chatwoot, incluindo a criação do webhook de saída que apontará para o seu Wuzapi.

Lembre-se de que o Wuzapi agora escuta na porta `8080` e o Traefik está configurado para rotear o tráfego para `wuzapi.zorbix.cloud` para esta porta. Certifique-se de que suas configurações de DNS e Traefik estejam corretas para que o domínio aponte para o seu serviço Wuzapi.
