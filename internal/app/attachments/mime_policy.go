package attachments

import "strings"

type MimePolicy struct {
	allowed map[string]struct{}
}

func NewMimePolicy(list []string) *MimePolicy {
	m := make(map[string]struct{}, len(list))
	for _, t := range list {
		if t == "" {
			continue
		}
		m[strings.ToLower(t)] = struct{}{}
	}

	return &MimePolicy{allowed: m}
}

func (p *MimePolicy) Allowed(mime string) bool {
	if p == nil {
		return true // additional check :)
	}

	_, isOk := p.allowed[strings.ToLower(mime)]
	return isOk
}
