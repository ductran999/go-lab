# OIDC

- Simple definition: Authentication(jwks + id_token) + Authorized (OAuth)
- Standard to make google, facebook, github, etc can expose same response structure to easy integrate with many other system.
- Support SSO for company app. (Gitlab, ArgoCD, Rancher, etc.)
- jwks allow key rotation. Required both old key and new key to ensure old valid token still can verified by app.
