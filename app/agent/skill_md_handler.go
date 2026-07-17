package agent

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// publicBaseURL is the only address agents outside the cluster can reach.
// The in-cluster Host header (e.g. core-mgmt.3cha.svc.cluster.local) is not
// reachable externally, so it must not be used here even though it's what
// requests to this handler arrive with.
const publicBaseURL = "https://api.nspublic.xyz/3cha"

// SkillMD serves a SKILL.md document (per the open Agent Skills spec,
// https://agentskills.io/specification) describing how an AI agent can call
// the 3cha Portal API directly instead of going through the web UI.
//
// It only covers read-only endpoints plus the stateless cipher utility —
// mutating endpoints (members/holidays/queue writes) are intentionally left
// out so an agent granted this skill can't accidentally change data.
func (h *handler) SkillMD(c *gin.Context) {
	c.Data(http.StatusOK, "text/markdown; charset=utf-8", []byte(buildSkillMD()))
}

func buildSkillMD() string {
	return strings.ReplaceAll(skillMDTemplate, "__BASE_URL__", publicBaseURL)
}

const skillMDTemplate = `---
name: 3cha-portal-api
description: Query 3cha Portal data over HTTP — today's daily-meeting duty queue, who's next in the rotation, queue history, team member list, and configured holidays — plus an AES-GCM encrypt/decrypt utility. Use when the user asks who's on duty / up next for the 3cha daily meeting, wants the member list or holiday schedule, or needs to encrypt/decrypt a value with 3cha's cipher endpoint. Read-only: does not create, update, or delete anything.
compatibility: Any HTTP client with network access to the 3cha Portal backend. No authentication is required — treat this as an internal-network tool and don't relay responses to untrusted parties.
metadata:
  scope: read-only
  source: 3cha-portal-backend
---

# 3cha Portal API

3cha Portal tracks a team's daily-meeting duty rotation (round-robin "queue"),
the member roster driving that rotation, and holidays that are skipped when
picking the next duty date. It also exposes a small AES-GCM cipher utility.

This skill covers only **read (GET) endpoints and the stateless cipher
endpoints**. The API also has write endpoints (create/delete members,
create/delete holidays, reset/mark-done/skip queue entries) — those are
deliberately not documented here so an agent using this skill can't mutate
data; use the web UI for those.

Base URL: __BASE_URL__

## Response envelope

Every endpoint returns the same JSON envelope:

` + "```json" + `
{
  "code": "SUCCESS",
  "message": "success",
  "data": { }
}
` + "```" + `

- ` + "`code`" + `: ` + "`SUCCESS`" + ` on success. Errors use ` + "`BAD_REQUEST`" + `, ` + "`NOT_FOUND`" + `, ` + "`CONFLICT`" + `, or ` + "`INTERNAL_ERROR`" + ` with a matching HTTP status.
- ` + "`data`" + ` is omitted (` + "`null`" + `) for errors and for some informational responses (see ` + "`/queue/today`" + ` below).
- Dates are ` + "`YYYY-MM-DD`" + ` strings, evaluated in the server's configured timezone (default ` + "`Asia/Bangkok`" + `).

## Endpoints

### 1. Today's duty — ` + "`GET /api/v1/queue/today`" + `

Who is on duty for today's daily meeting, their position in the rotation, and
who's next.

` + "```bash" + `
curl __BASE_URL__/api/v1/queue/today
` + "```" + `

Normal response (` + "`data`" + `):

` + "```json" + `
{
  "id": "b3f1...",
  "queueDate": "2026-07-17",
  "status": "pending",
  "memberId": "a1c2...",
  "memberName": "Nawaphong",
  "avatarColor": "#6366f1",
  "confluenceUrl": "https://...",
  "position": 2,
  "totalMembers": 5,
  "next": {
    "memberName": "Somchai",
    "avatarColor": "#22c55e",
    "queueDate": "2026-07-18"
  }
}
` + "```" + `

- ` + "`status`" + ` is one of ` + "`pending`" + `, ` + "`done`" + `, ` + "`skipped`" + `.
- ` + "`next`" + ` is omitted if there's no active member or no upcoming working day found.
- If today is a holiday / weekend and there's no meeting, the response is
  still HTTP 200 but with ` + "`code: \"NO_MEETING_TODAY\"`" + ` and no ` + "`data`" + `:

` + "```json" + `
{ "code": "NO_MEETING_TODAY", "message": "วันนี้ไม่มี daily meeting (วันหยุด)" }
` + "```" + `

### 2. Queue history — ` + "`GET /api/v1/queue`" + `

The last 30 days of queue entries, newest first. No query parameters (the
window is fixed).

` + "```bash" + `
curl __BASE_URL__/api/v1/queue
` + "```" + `

` + "```json" + `
[
  { "id": "...", "queueDate": "2026-07-17", "status": "pending", "memberId": "...", "memberName": "Nawaphong", "avatarColor": "#6366f1" }
]
` + "```" + `

### 3. Members — ` + "`GET /api/v1/members`" + `

All team members (active and inactive), ordered by ` + "`sortOrder`" + ` then
` + "`createdAt`" + `. This is the order the round-robin rotation follows.

` + "```bash" + `
curl __BASE_URL__/api/v1/members
` + "```" + `

` + "```json" + `
[
  { "id": "...", "name": "Nawaphong", "avatarColor": "#6366f1", "isActive": true, "sortOrder": 0, "createdAt": "...", "updatedAt": "..." }
]
` + "```" + `

### 4. Holidays — ` + "`GET /api/v1/holidays`" + `

Configured holidays (days the rotation skips), newest first.

` + "```bash" + `
curl __BASE_URL__/api/v1/holidays
` + "```" + `

` + "```json" + `
[
  { "id": "...", "holidayDate": "2026-12-31", "name": "New Year's Eve", "createdAt": "..." }
]
` + "```" + `

### Handling the cipher key

The API never generates or stores keys — the caller supplies one on every
call. Before your first ` + "`encrypt`" + `/` + "`decrypt`" + ` call in a
conversation:

1. Ask the user for the key value.
2. Ask whether they'd like it remembered for next time, and under what label
   (e.g. ` + "`work-notes-key`" + `, or ` + "`default`" + ` if they don't care).
3. If they say yes and your runtime has a persistent memory/notes mechanism,
   save the key under that label so future sessions can reuse it without
   asking again.
4. On later calls, check for a saved key under the label the user names (or
   ` + "`default`" + ` if unspecified) before asking again.

**Never invent, guess, or silently reuse a key from a different label.** A
new or different key must always come from the user explicitly — if nothing
is saved under the requested label, ask.

### 5. Encrypt — ` + "`POST /api/v1/cipher/encrypt`" + `

AES-GCM encrypt with a caller-supplied key (128/192/256-bit, i.e. 16/24/32
ASCII characters). Output is hex-encoded (random 12-byte nonce prepended to
the ciphertext). See "Handling the cipher key" above before calling this.

` + "```bash" + `
curl -X POST __BASE_URL__/api/v1/cipher/encrypt \
  -H 'Content-Type: application/json' \
  -d '{"text":"hello world","key":"0123456789abcdef"}'
` + "```" + `

` + "```json" + `
{ "code": "SUCCESS", "message": "success", "data": { "result": "9f2a...hex..." } }
` + "```" + `

### 6. Decrypt — ` + "`POST /api/v1/cipher/decrypt`" + `

Same shape, reversed: ` + "`text`" + ` is the hex string from ` + "`encrypt`" + `, ` + "`key`" + `
must match the key used to encrypt it. See "Handling the cipher key" above
before calling this.

` + "```bash" + `
curl -X POST __BASE_URL__/api/v1/cipher/decrypt \
  -H 'Content-Type: application/json' \
  -d '{"text":"9f2a...hex...","key":"0123456789abcdef"}'
` + "```" + `

` + "```json" + `
{ "code": "SUCCESS", "message": "success", "data": { "result": "hello world" } }
` + "```" + `

## Installing this as a skill

If your agent loads skills from a local directory rather than fetching a URL
at runtime, save this document's content as ` + "`3cha-portal-api/SKILL.md`" + `
(the folder name must match the ` + "`name`" + ` field above) inside your skills
directory:

` + "```bash" + `
mkdir -p 3cha-portal-api
curl __BASE_URL__/api/v1/agent/skill-md -o 3cha-portal-api/SKILL.md
` + "```" + `
`
