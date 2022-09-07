package utils

const (
	CardTypeVisa       = "VISA"
	CardTypeMasterCard = "MASTERCARD"
)

func GetCardTypeByMask(cardMask string) string {
	switch cardMask[0:1] {
	case "4":
		return "VISA"
	case "5":
		return "MASTERCARD"
	default:
		return "UNDEFINED"
	}
}

func GetCardLastDigitsByMask(cardMask string) string {
	lastDigits := "0000"
	if cardMask != "" {
		lastDigits = cardMask[len(cardMask)-4:]
	}
	return lastDigits
}
