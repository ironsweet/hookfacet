package hookfacet

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEventJSONShape(t *testing.T) {
	ev := Event{SessionID: "s1", Kind: KindTurnEnd, Payload: Payload{}}
	b, _ := json.Marshal(ev)
	assert.JSONEq(t, `{"session_id":"s1","kind":"turn_end","payload":{}}`, string(b))
}

func TestDecisionJSONShape(t *testing.T) {
	b, _ := json.Marshal(Allow())
	assert.JSONEq(t, `{}`, string(b))

	b, _ = json.Marshal(Block("not satisfied"))
	assert.JSONEq(t, `{"block":true,"reason":"not satisfied"}`, string(b))
}
