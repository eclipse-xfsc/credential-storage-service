package model

import "github.com/eclipse-xfsc/oid4-vci-vp-library/model/presentation"

type GetCredentialModel struct {
	Credentials map[string]interface{}      `json:"credentials,omitempty"`
	Receipt     string                      `json:"receipt,omitempty"`
	Groups      []presentation.FilterResult `json:"groups,omitempty"`
}
