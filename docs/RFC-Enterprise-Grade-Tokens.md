# RFC: Enterprise-Grade Token Authentication

**Status:** Draft
**Author:** Gemini
**Date:** 2025-11-07

## 1. Summary

This RFC proposes a series of enhancements to our token-based authentication system (access and refresh tokens) to align with enterprise-grade security and scalability standards. The core proposals are:

1.  Transition to using secure, `HttpOnly` cookies for refresh token transport.
2.  Implement refresh token rotation and a robust server-side invalidation mechanism.
3.  Adopt lightweight, stateless access tokens containing minimal claims.

These changes will significantly harden our system against common web vulnerabilities like Cross-Site Scripting (XSS) and token hijacking, while also improving performance and maintainability.

## 2. Motivation

Our current token implementation, while functional, exposes the application to unnecessary risks and limitations that are unacceptable at an enterprise scale:

*   **XSS Vulnerability:** By returning the refresh token in the API response body, we force clients (especially web browsers) to store it in JavaScript-accessible storage (e.g., `localStorage`). This makes it highly vulnerable to theft via XSS attacks. A stolen refresh token can lead to complete, long-term account takeover.
*   **No Robust Invalidation:** A compromised refresh token can be used to generate new access tokens until it expires. We lack a reliable mechanism to immediately invalidate a specific user's session on the server-side in response to a security event (e.g., password change, logout, detected intrusion).
*   **Stale Data & Token Bloat:** Including user roles and profile information directly within the access token (JWT) leads to two problems: the data can become stale if the user's profile changes, and the token size increases, adding unnecessary overhead to every single API request.

To be considered enterprise-grade, our authentication system must prioritize security and adhere to modern best practices that mitigate these specific risks.

## 3. Detailed Design

This proposal is broken down into three key enhancements that build upon each other.

### 3.1. Secure Refresh Token Handling via HttpOnly Cookies

We will fundamentally change how refresh tokens are transported and stored on the client.

*   **New Flow:**
    1.  The `/login` endpoint will no longer include the refresh token in the JSON response. Instead, it will set the refresh token in a cookie with the following critical attributes:
        *   **`HttpOnly`**: Prevents any JavaScript running on the client from accessing the cookie. This is our primary defense against XSS-based token theft.
        *   **`Secure=true`**: Ensures the cookie is only ever sent over an encrypted (HTTPS) connection.
        *   **`SameSite=Strict`**: Provides strong protection against Cross-Site Request Forgery (CSRF) by preventing the browser from sending the cookie along with cross-site requests.
        *   **`Path=/v2/tokens`**: Scopes the cookie so that the browser only sends it to our token refresh endpoint, minimizing its exposure.
    2.  The `/tokens` (refresh) endpoint will now read the refresh token from this incoming cookie instead of from a request body or header.
    3.  The access token will continue to be returned in the JSON body, as it is short-lived and needed by client-side code to make API requests.

### 3.2. Refresh Token Rotation and Server-Side Invalidation

To limit the utility of any potentially compromised refresh token, we will implement rotation and enhance our database schema for immediate invalidation.

*   **Token Rotation:** Each time a refresh token is used successfully at the `/tokens` endpoint, it will be immediately invalidated, and a **new** refresh token will be issued and set as a new `HttpOnly` cookie. If an attacker ever tries to reuse an old (already used) refresh token, the server will detect it, reject the request, and can be configured to automatically invalidate the entire family of tokens for that user session, forcing a logout on all devices as a security precaution.

*   **Database Schema (`user_tokens` table):** The schema will be updated to support this flow. We will store a record for every active refresh token.

    ```sql
    CREATE TABLE user_tokens (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        token_hash VARCHAR(255) NOT NULL UNIQUE, -- A SHA-256 hash of the refresh token
        family VARCHAR(255) NOT NULL, -- A value to link a chain of rotated tokens to an initial session
        expires_at TIMESTAMPTZ NOT NULL,
        is_revoked BOOLEAN NOT NULL DEFAULT false,
        created_at TIMESTAMPTZ NOT NULL DEFAULT now()
    );
    CREATE INDEX idx_user_tokens_user_id ON user_tokens(user_id);
    ```
    *   **`token_hash`**: We store a secure hash of the token, not the raw token itself. This prevents session hijacking even if the database is compromised.
    *   **`family`**: A unique identifier for the initial login session. If a rotated token is reused, we can revoke all tokens in the same family.
    *   **`is_revoked`**: A boolean flag to explicitly invalidate a token (e.g., on logout or password change).

### 3.3. Lightweight, Stateless Access Tokens

Access tokens (JWTs) should be small, short-lived, and contain only the information required for identity verification and locating permissions, not the permissions themselves.

*   **Minimal Claims:** The JWT payload should be reduced to include only essential, immutable claims:
    *   `sub` (Subject): The user's unique and stable ID.
    *   `exp` (Expiration Time): A short lifetime (e.g., 15 minutes).
    *   `iat` (Issued At): The time the token was issued.
    *   `iss` (Issuer): The service that issued the token (e.g., `go-community-api`).
    *   `aud` (Audience): The service(s) intended to consume the token.

*   **Fetching Fresh Data:** When a service receives a request, it should use the `sub` (user ID) from the validated token to fetch the user's most current roles, permissions, and profile data from a fast cache (like Redis) or directly from the database. This completely solves the "stale data" problem and keeps tokens lean and performant. This pattern is a prerequisite for the centralized ReBAC model.

## 4. API Changes

*   **`POST /v2/users/login`**
    *   **Response Body:** Will contain the `access_token` and user profile information.
    *   **Response Headers:** Will include a `Set-Cookie` header for the new refresh token.
        ```
        Set-Cookie: rt=...; HttpOnly; Secure; SameSite=Strict; Path=/v2/tokens; Max-Age=2592000
        ```

*   **`POST /v2/tokens`** (Method should be changed from GET to POST)
    *   **Request:** Will not have a body. It will expect the refresh token in the `rt` cookie sent by the browser.
    *   **Response Body:** Will contain the new `access_token`.
    *   **Response Headers:** Will include a `Set-Cookie` header with the **new** rotated refresh token.

## 5. Adoption Strategy

1.  **Database Migration:** Introduce the new `user_tokens` table schema and migrate any existing token data.
2.  **Update Token Logic:** Modify the `/login` and `/tokens` use cases to generate lightweight access tokens and rotated refresh tokens, storing their hashes in the new table.
3.  **Introduce Cookie Flow:** Update the `/login` and `/tokens` handlers to set/read the `HttpOnly` cookies. This can be rolled out behind a feature flag to avoid breaking existing clients.
4.  **Client-Side Refactor:** Update frontend clients to stop storing the refresh token in `localStorage` and to handle the new cookie-based flow for refreshing sessions.
5.  **Deprecate Old Flow:** Once all clients are migrated, remove the old token handling logic and remove refresh tokens from the API body for good.
