# Keycloak Playground

## Multi Tenant Support

Solution: B2B

- 1 Tenant = 1 Realm:

Pros:

- allow custom logo
- integrate customer company LDAP
- isolate data

Cons:

- Performance low

## Well-knowns: Keycloak HATEOAS tiệm cận vì đây là trang tĩnh.

- `issuer`: domain iss the `id_token`
- `authorization_endpoint`: đường dẫn đến trang đăng nhập.
- `token_endpoint`: the exchange token url
- `jwks_uri`: public key
