package bluetooth

// LookupManufacturer returns a human-readable name for a Bluetooth SIG company ID.
// See: https://www.bluetooth.com/specifications/assigned-numbers/
func LookupManufacturer(companyID uint16) string {
	if name, ok := companyNames[companyID]; ok {
		return name
	}
	return ""
}

var companyNames = map[uint16]string{
	0x004C: "Apple",
	0x0006: "Microsoft",
	0x00E0: "Google",
	0x0075: "Samsung",
	0x0310: "Xiaomi",
	0x0157: "Huawei",
	0x038F: "Garmin",
	0x0087: "Bose",
	0x012D: "Sony",
	0x00D2: "LG",
	0x0171: "Amazon",
	0x02FF: "Tile",
	0x0059: "Nordic",
	0x000D: "Texas Inst.",
	0x0822: "Tuya/Govee",
	0x0131: "JBL",
	0x00E3: "Harman",
	0x0002: "Intel",
	0x000F: "Broadcom",
	0x000A: "Qualcomm",
	0x0499: "Ruuvi",
	0x0672: "Shenzhen",
	0x015D: "Espressif",
	0x0047: "Plantronics",
	0x01DA: "Jabra",
	0x0056: "Sony Erics.",
	0x0078: "Nike",
	0x0154: "Belkin",
	0x00AA: "Realtek",
	0x048F: "Wyze",
	0x0958: "IKEA",
	0x09A7: "Ring",
	0x0246: "Logitech",
	0x0060: "Motorola",
	0x03DA: "Fitbit",
	0x0988: "Sonos",
	0x0269: "Oura",
	0x0473: "Withings",
	0x02A9: "Anker",
	0x0397: "TP-Link",
	0x0362: "Yeelight",
}
