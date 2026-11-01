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
package create

import (
	"encoding/json"
	"strings"
	"testing"

	"loxicmd/pkg/api"
)

func TestSelectToNumQuicCid(t *testing.T) {
	if got := SelectToNum(SelectQuicCid); got != 7 {
		t.Errorf("SelectToNum(%q) = %d, want 7", SelectQuicCid, got)
	}
	// The other values, and the default for an unknown one, are what they were
	for sel, want := range map[string]int{"rr": 0, "hash": 1, "priority": 2, "persist": 3, "lc": 4, "n2": 5, "n3": 6, "other": 0} {
		if got := SelectToNum(sel); got != want {
			t.Errorf("SelectToNum(%q) = %d, want %d", sel, got, want)
		}
	}
}

func TestQuicLbArgs(t *testing.T) {
	tests := []struct {
		name      string
		sel       string
		cr        uint8
		sidLen    uint8
		crSet     bool
		sidLenSet bool
		want      *api.LbQuic
		wantErr   string
	}{
		{name: "defaults", sel: SelectQuicCid, sidLen: 4, want: &api.LbQuic{SidLen: 4}},
		{name: "cr 7", sel: SelectQuicCid, sidLen: 4, cr: 7, crSet: true, want: &api.LbQuic{SidLen: 4, Cr: 7}},
		{name: "sidlen 4 named", sel: SelectQuicCid, sidLen: 4, sidLenSet: true, want: &api.LbQuic{SidLen: 4}},
		{name: "cr 8", sel: SelectQuicCid, sidLen: 4, cr: 8, crSet: true, wantErr: "--quic-cid-cr must be"},
		{name: "sidlen 0", sel: SelectQuicCid, sidLenSet: true, wantErr: "--quic-cid-sidlen must be"},
		{name: "sidlen 5", sel: SelectQuicCid, sidLen: 5, sidLenSet: true, wantErr: "--quic-cid-sidlen must be"},
		{name: "other selector, no flags", sel: "hash", sidLen: 4},
		{name: "other selector, cr", sel: "hash", sidLen: 4, cr: 1, crSet: true, wantErr: "need --select=quic-cid"},
		{name: "other selector, cr 0 named", sel: "rr", sidLen: 4, crSet: true, wantErr: "need --select=quic-cid"},
		{name: "other selector, sidlen", sel: "rr", sidLen: 4, sidLenSet: true, wantErr: "need --select=quic-cid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := CreateLoadBalancerOptions{Select: tt.sel, QuicCidCr: tt.cr, QuicCidSidLen: tt.sidLen}
			got, err := quicLbArgs(&o, tt.crSet, tt.sidLenSet)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want one with %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("err = %v", err)
			}
			if (got == nil) != (tt.want == nil) || (got != nil && *got != *tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

// A rule of any other selector must send the bytes it sent before quicLb existed.
func TestOtherSelectorRequestHasNoQuicLb(t *testing.T) {
	b, err := json.Marshal(api.LoadBalancerService{ExternalIP: "192.0.2.1", Port: 80, Protocol: "tcp", Sel: 1})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "quicLb") {
		t.Errorf("request has quicLb: %s", b)
	}
}
