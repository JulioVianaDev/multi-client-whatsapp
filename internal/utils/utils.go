package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"multi-client-whatsapp/internal/types"

	whatsappTypes "go.mau.fi/whatsmeow/types"
)

// GenerateInstanceKey generates a unique instance key
func GenerateInstanceKey() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ParseJIDWithLIDSupport parses a JID with support for both @s.whatsapp.net and @lid
func ParseJIDWithLIDSupport(phone string, instance *types.Instance) (whatsappTypes.JID, error) {
	// First try to parse as regular JID
	if jid, err := whatsappTypes.ParseJID(phone); err == nil {
		return jid, nil
	}

	// Check if it's a LID format (ends with @lid)
	if strings.HasSuffix(phone, "@lid") {
		// Try to parse as LID
		lidJID, err := whatsappTypes.ParseJID(phone)
		if err != nil {
			return whatsappTypes.JID{}, fmt.Errorf("invalid LID format: %v", err)
		}

		// Try to get phone number for this LID
		if instance.Client != nil && instance.Client.Store != nil && instance.Client.Store.LIDs != nil {
			ctx := context.Background()
			pn, err := instance.Client.Store.LIDs.GetPNForLID(ctx, lidJID)
			if err != nil {
				log.Printf("Warning: Could not get phone number for LID %s: %v", lidJID.String(), err)
				// Return the LID anyway, as it might still work
				return lidJID, nil
			}
			if !pn.IsEmpty() {
				log.Printf("Resolved LID %s to phone number %s", lidJID.String(), pn.String())
				return pn, nil
			}
		}

		// Return the LID if we can't resolve it
		return lidJID, nil
	}

	// If it doesn't end with @s.whatsapp.net or @lid, try to add @s.whatsapp.net
	if !strings.Contains(phone, "@") {
		phoneWithSuffix := phone + "@s.whatsapp.net"
		if jid, err := whatsappTypes.ParseJID(phoneWithSuffix); err == nil {
			return jid, nil
		}
	}

	return whatsappTypes.JID{}, fmt.Errorf("invalid phone number format: %s", phone)
}
