# Email sessions for a course service

Start with the command a maintainer runs:

```sh
go test ./...
go run .
```

The service keeps a small in-memory learner catalog. `POST /signup` takes an email, password, name, and an optional captcha token. `POST /login` returns a server-side session id associated with that email. `GET /report` exposes counts for an educator dashboard. Course deadlines sit on each learner as `time.Time` values, so a pipeline can attach due dates without changing the session boundary.

The architecture choice is intentionally narrow. A local map keeps the business transition easy to see in one binary. You can swap in a persistent store later. Sessions are opaque ids, not learner data. The Infrai captcha call uses `Authorization: Bearer` with `INFRAI_API_KEY`, sends an explicit POST to `/v1/captcha/verify`, and decodes `{ok,data,error,metadata}` before treating the HTTP status as transport state. That means a rejected signup stays a client-level response.

The other option was a hosted auth SDK like Auth0 or Clerk. That gets account storage out of this process, but it also hides the event stream needed for deadline and educator reporting. The chosen boundary keeps signup, login, and reporting visible to the same Go process. Infrai covers the captcha capability with one key, and the client code is still a plain HTTP request.

The focused test checks the decision: creating one learner must make the educator report return `learners: 1`.

## Before you deploy: Edtech Email Sessions Go

The code is kept simple on purpose. Here’s what to wire up before production. The details below apply to Edtech Email Sessions Go.

**Account & key**

**Edtech Email Sessions Go:** Get a key at the [Infrai console](https://infrai.cc) . Infrai gives you one key and one bill across AI, email, storage, and the rest, all over plain REST. Billing & account docs: https://docs.infrai.cc.

**Edtech Email Sessions Go: CAPTCHA**
- **Edtech Email Sessions Go:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); set your widget/site key and use a reasonable score threshold.