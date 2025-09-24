package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	"multi-client-whatsapp/internal/types"

	whatsappTypes "go.mau.fi/whatsmeow/types"
)

// ValidateBrazilianPhoneNumber validates and corrects Brazilian phone numbers
func ValidateBrazilianPhoneNumber(phone string, instance *types.Instance) (string, bool, error) {
	// Remove any non-digit characters
	cleanPhone := strings.ReplaceAll(phone, "@s.whatsapp.net", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "@lid", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "(", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, ")", "")

	// Check if it's a Brazilian number (starts with 55)
	if !strings.HasPrefix(cleanPhone, "55") {
		return phone, false, fmt.Errorf("not a Brazilian phone number")
	}

	// Extract the area code and number
	if len(cleanPhone) < 12 {
		return phone, false, fmt.Errorf("invalid phone number length")
	}

	areaCode := cleanPhone[2:4]
	number := cleanPhone[4:]

	// Check if it's a mobile number (area codes that use 9-digit numbers)
	mobileAreaCodes := map[string]bool{
		"11": true, "12": true, "13": true, "14": true, "15": true, "16": true, "17": true, "18": true, "19": true,
		"21": true, "22": true, "24": true, "27": true, "28": true,
		"31": true, "32": true, "33": true, "34": true, "35": true, "37": true, "38": true,
		"41": true, "42": true, "43": true, "44": true, "45": true, "46": true, "47": true, "48": true, "49": true,
		"51": true, "53": true, "54": true, "55": true,
		"61": true, "62": true, "63": true, "64": true, "65": true, "66": true, "67": true, "68": true, "69": true,
		"71": true, "73": true, "74": true, "75": true, "77": true, "79": true,
		"81": true, "82": true, "83": true, "84": true, "85": true, "86": true, "87": true, "88": true, "89": true,
		"91": true, "92": true, "93": true, "94": true, "95": true, "96": true, "97": true, "98": true, "99": true,
	}

	isMobileArea := mobileAreaCodes[areaCode]

	// First, try the number as received (original format)
	originalPhone := "55" + areaCode + number + "@s.whatsapp.net"
	log.Printf("🔍 Step 1: Checking original number: %s", originalPhone)
	originalExists, _ := CheckPhoneExists(originalPhone, instance)
	log.Printf("📊 Original number %s exists: %v", originalPhone, originalExists)

	if isMobileArea {
		// For mobile numbers, try variations
		if len(number) == 8 {
			// Try adding 9 at the beginning
			phoneWith9 := "55" + areaCode + "9" + number + "@s.whatsapp.net"
			with9Exists, _ := CheckPhoneExists(phoneWith9, instance)

			// Return the one that exists, or original if neither exists
			if originalExists {
				return originalPhone, true, nil
			} else if with9Exists {
				return phoneWith9, true, nil
			} else {
				return originalPhone, false, nil
			}
		} else if len(number) == 9 {
			// Check if it starts with 9
			if strings.HasPrefix(number, "9") {
				// Try removing the 9
				phoneWithout9 := "55" + areaCode + number[1:] + "@s.whatsapp.net"
				without9Exists, _ := CheckPhoneExists(phoneWithout9, instance)

				// Return the one that exists, or original if neither exists
				if originalExists {
					return originalPhone, true, nil
				} else if without9Exists {
					return phoneWithout9, true, nil
				} else {
					return originalPhone, false, nil
				}
			} else {
				// Number has 9 digits but doesn't start with 9
				// Try adding 9 at the beginning
				phoneWith9 := "55" + areaCode + "9" + number + "@s.whatsapp.net"
				with9Exists, _ := CheckPhoneExists(phoneWith9, instance)

				// Return the one that exists, or original if neither exists
				if originalExists {
					return originalPhone, true, nil
				} else if with9Exists {
					return phoneWith9, true, nil
				} else {
					return originalPhone, false, nil
				}
			}
		} else if len(number) == 10 {
			// Number has 10 digits, might need 9 added
			phoneWith9 := "55" + areaCode + "9" + number + "@s.whatsapp.net"
			with9Exists, _ := CheckPhoneExists(phoneWith9, instance)

			// Return the one that exists, or original if neither exists
			if originalExists {
				return originalPhone, true, nil
			} else if with9Exists {
				return phoneWith9, true, nil
			} else {
				return originalPhone, false, nil
			}
		} else if len(number) == 11 {
			// Number has 11 digits, might need 9 removed (like 5541991968071 -> 554191968071)
			// Check if it starts with 9
			if strings.HasPrefix(number, "9") {
				// Try removing the 9
				phoneWithout9 := "55" + areaCode + number[1:] + "@s.whatsapp.net"
				log.Printf("🔍 Step 2: Checking without 9: %s", phoneWithout9)
				without9Exists, _ := CheckPhoneExists(phoneWithout9, instance)
				log.Printf("📊 Without 9 number %s exists: %v", phoneWithout9, without9Exists)

				// Return the one that exists, or original if neither exists
				if originalExists {
					return originalPhone, true, nil
				} else if without9Exists {
					return phoneWithout9, true, nil
				} else {
					return originalPhone, false, nil
				}
			} else {
				// Number has 11 digits but doesn't start with 9
				// Try adding 9 at the beginning
				phoneWith9 := "55" + areaCode + "9" + number + "@s.whatsapp.net"
				with9Exists, _ := CheckPhoneExists(phoneWith9, instance)

				// Return the one that exists, or original if neither exists
				if originalExists {
					return originalPhone, true, nil
				} else if with9Exists {
					return phoneWith9, true, nil
				} else {
					return originalPhone, false, nil
				}
			}
		}
	}

	// For landline numbers or any other case, return the original
	return originalPhone, originalExists, nil
}

// CheckPhoneExists checks if a phone number exists on WhatsApp
func CheckPhoneExists(phone string, instance *types.Instance) (bool, error) {
	if instance == nil || instance.Client == nil {
		return false, fmt.Errorf("instance not available")
	}

	// Only check if the JID is a user with @s.whatsapp.net
	if strings.Contains(phone, "@s.whatsapp.net") {
		log.Printf("🔍 Checking if phone exists: %s", phone)

		// First try the IsOnWhatsApp API
		data, err := instance.Client.IsOnWhatsApp([]string{phone})
		if err != nil {
			log.Printf("❌ Error checking phone %s: %v", phone, err)
			return false, fmt.Errorf("failed to check if user is on whatsapp: %v", err)
		}

		log.Printf("📊 WhatsApp API response for %s: %+v", phone, data)

		// Check if any number exists according to the API and detect redirections
		apiExists := false
		redirectedJID := ""
		for _, v := range data {
			log.Printf("📱 Phone %s - IsIn: %v", v.JID, v.IsIn)
			if v.IsIn {
				apiExists = true
				// Check if WhatsApp redirected us to a different JID
				if v.JID.String() != phone {
					redirectedJID = v.JID.String()
					log.Printf("🔄 WhatsApp redirected %s to %s", phone, v.JID.String())
				}
			}
		}

		// If API says it doesn't exist, return false
		if !apiExists {
			log.Printf("❌ Phone %s does NOT exist on WhatsApp (API check)", phone)
			return false, nil
		}

		// If WhatsApp redirected us to a different JID, that means the original doesn't exist
		if redirectedJID != "" {
			log.Printf("❌ Phone %s does NOT exist on WhatsApp (redirected to %s)", phone, redirectedJID)
			return false, nil
		}

		// If API says it exists and no redirection, do a double-check by trying to get user info
		jid, err := whatsappTypes.ParseJID(phone)
		if err != nil {
			log.Printf("❌ Error parsing JID %s: %v", phone, err)
			return false, err
		}

		// Try to get user info - this will fail if the user doesn't exist
		userInfo, err := instance.Client.GetUserInfo([]whatsappTypes.JID{jid})
		if err != nil {
			log.Printf("❌ Phone %s does NOT exist on WhatsApp (GetUserInfo failed): %v", phone, err)
			return false, nil
		}

		if len(userInfo) == 0 {
			log.Printf("❌ Phone %s does NOT exist on WhatsApp (no user info)", phone)
			return false, nil
		}

		log.Printf("✅ Phone %s exists on WhatsApp (confirmed by GetUserInfo)", phone)
		return true, nil
	}

	return false, fmt.Errorf("phone number must end with @s.whatsapp.net for validation")
}

// ValidateAndCorrectPhone validates and corrects a phone number for Brazilian numbers
func ValidateAndCorrectPhone(phone string, instance *types.Instance) (string, error) {
	// First try to parse as regular JID
	if jid, err := whatsappTypes.ParseJID(phone); err == nil {
		// If it's already a valid JID, check if it exists
		if exists, _ := CheckPhoneExists(jid.String(), instance); exists {
			return jid.String(), nil
		}
	}

	// Check if it's a LID format (ends with @lid)
	if strings.HasSuffix(phone, "@lid") {
		lidJID, err := whatsappTypes.ParseJID(phone)
		if err != nil {
			return phone, fmt.Errorf("invalid LID format: %v", err)
		}

		// Try to get phone number for this LID
		if instance.Client != nil && instance.Client.Store != nil && instance.Client.Store.LIDs != nil {
			ctx := context.Background()
			pn, err := instance.Client.Store.LIDs.GetPNForLID(ctx, lidJID)
			if err != nil {
				log.Printf("Warning: Could not get phone number for LID %s: %v", lidJID.String(), err)
				return lidJID.String(), nil
			}
			if !pn.IsEmpty() {
				log.Printf("Resolved LID %s to phone number %s", lidJID.String(), pn.String())
				return pn.String(), nil
			}
		}

		return lidJID.String(), nil
	}

	// If it doesn't end with @s.whatsapp.net or @lid, try Brazilian validation
	if !strings.Contains(phone, "@") {
		// Try Brazilian phone validation
		validPhone, exists, err := ValidateBrazilianPhoneNumber(phone, instance)
		if err == nil {
			if exists {
				log.Printf("Validated Brazilian phone number: %s -> %s (exists)", phone, validPhone)
				return validPhone, nil
			} else {
				log.Printf("Validated Brazilian phone number: %s -> %s (doesn't exist)", phone, validPhone)
				return validPhone, nil
			}
		}

		// If Brazilian validation fails, try adding @s.whatsapp.net
		phoneWithSuffix := phone + "@s.whatsapp.net"
		if jid, err := whatsappTypes.ParseJID(phoneWithSuffix); err == nil {
			return jid.String(), nil
		}
	} else if strings.Contains(phone, "@s.whatsapp.net") {
		// If it already has @s.whatsapp.net, try Brazilian validation on the number part
		numberPart := strings.ReplaceAll(phone, "@s.whatsapp.net", "")
		validPhone, exists, err := ValidateBrazilianPhoneNumber(numberPart, instance)
		if err == nil {
			if exists {
				log.Printf("Validated Brazilian phone number with suffix: %s -> %s (exists)", phone, validPhone)
				return validPhone, nil
			} else {
				log.Printf("Validated Brazilian phone number with suffix: %s -> %s (doesn't exist)", phone, validPhone)
				return validPhone, nil
			}
		}
	}

	return phone, fmt.Errorf("invalid phone number format: %s", phone)
}
