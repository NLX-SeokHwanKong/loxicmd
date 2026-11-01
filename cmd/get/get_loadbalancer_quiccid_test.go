/*
 * Copyright (c) 2026 NetLOX Inc
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at:
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package get

import (
	"encoding/json"
	"testing"

	"loxicmd/pkg/api"
)

func TestSelectCell(t *testing.T) {
	tests := []struct {
		name    string
		service api.LoadBalancerService
		want    string
	}{
		{"quic-cid", api.LoadBalancerService{Sel: 7, QuicLb: &api.LbQuic{SidLen: 4, Cr: 2}}, "quic-cid:sidlen=4,cr=2"},
		{"quic-cid without parameters", api.LoadBalancerService{Sel: 7}, "quic-cid"},
		{"hash", api.LoadBalancerService{Sel: 1}, "hash"},
		{"rr", api.LoadBalancerService{Sel: 0}, "rr"},
		{"unknown", api.LoadBalancerService{Sel: 99}, "rr"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := selectCell(tt.service); got != tt.want {
				t.Errorf("selectCell = %q, want %q", got, tt.want)
			}
		})
	}
}

// What save writes comes from this model: the parameters must survive it, and a file without them must still load.
func TestSavedRuleKeepsQuicLb(t *testing.T) {
	var in api.LbRuleModGet
	body := `{"lbAttr":[{"serviceArguments":{"externalIP":"192.0.2.1","port":443,"protocol":"udp","sel":7,"quicLb":{"sidLen":4,"cr":3}}}]}`
	if err := json.Unmarshal([]byte(body), &in); err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out api.LbRuleModGet
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	q := out.LbRules[0].Service.QuicLb
	if q == nil || q.SidLen != 4 || q.Cr != 3 {
		t.Errorf("after save: %+v in %s", q, b)
	}

	var old api.LbRuleModGet
	if err := json.Unmarshal([]byte(`{"lbAttr":[{"serviceArguments":{"externalIP":"192.0.2.1","port":80,"protocol":"tcp","sel":1}}]}`), &old); err != nil {
		t.Fatal(err)
	}
	if old.LbRules[0].Service.QuicLb != nil {
		t.Errorf("a file without quicLb gained it: %+v", old.LbRules[0].Service.QuicLb)
	}
}
