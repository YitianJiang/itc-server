package kitc

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"code.byted.org/gopkg/logs"
)

// When tlb support user-defined routing dyeing, this file will be deprecated

type tmp_ddp_ctx_key struct{}

var ddp_key tmp_ddp_ctx_key

type tmp_ddp_ctx_val struct {
	Uid string
	Did string
}

type DDPRule struct {
	Uid        string `json:"uid,omitempty"`
	Did        string `json:"did,omitempty"`
	DDPId      int64  `json:"ddp_id"`
	ValidUntil int64  `json:"valid_until"`
}

func (r *DDPRule) Match(user *tmp_ddp_ctx_val, now int64) bool {
	check_uid := len(r.Uid) > 0
	check_did := len(r.Did) > 0

	if r.ValidUntil < now {
		return false
	}

	if !check_uid && !check_did {
		return false
	}

	if (check_uid && r.Uid != user.Uid) || (check_did && r.Did != user.Did) {
		return false
	}

	return true
}

func WithUIDAndDIDSet(ctx context.Context) bool {
	return ctx.Value(ddp_key) != nil
}

func NewCtxWithUIDAndDID(ctx context.Context, uid int64, did int64) context.Context {
	value := &tmp_ddp_ctx_val{
		Uid: strconv.FormatInt(uid, 10),
		Did: strconv.FormatInt(did, 10),
	}

	return context.WithValue(ctx, ddp_key, value)
}

func (kc *KitcClient) checkDDPRoutingRules(ctx context.Context, r RPCMeta) (string, bool) {
	tmp := ctx.Value(ddp_key)

	if value, ok := tmp.(*tmp_ddp_ctx_val); !ok {
		logs.Warnf("KITC: invalid ddp value set: %v", value)
		return "", false
	} else {
		if len(value.Uid) == 0 && len(value.Did) == 0 {
			return "", false
		}

		rules, err := kc.remoteConfiger.GetDDPRules(r)
		if err != nil {
			logs.Warnf("KITC: %s", err.Error())
			return "", false
		}

		now := time.Now().UnixNano()
		ids := make([]string, 0, 10)
		for _, rule := range *rules {
			if rule.Match(value, now) {
				ids = append(ids, strconv.FormatInt(rule.DDPId, 10))
			}
		}

		if len(ids) == 0 {
			return "", true
		}

		sort.Strings(ids)
		return strings.Join(ids, ","), true
	}
}

func (rc *remoteConfiger) GetDDPRules(r RPCMeta) (*[]DDPRule, error) {
	buf := make([]byte, 0, 100)
	buf = append(buf, r.From...)
	buf = append(buf, '/')
	buf = append(buf, r.FromCluster...)

	rules := make([]DDPRule, 0)

	key := path.Join("/kite/ddp", string(buf))
	val, err := rc.kvstorer.Get(key)
	if err != nil {
		return &rules, nil
	}

	if err := json.Unmarshal([]byte(val), &rules); err != nil {
		return nil, fmt.Errorf("invalid ddp setting: %s", val)
	}

	return &rules, nil
}
