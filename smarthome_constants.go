package go_fritzbox_api

// Beschreibung inkl. Fehler aus der Doku übernommen ;)
const (
	CHanfun           = "HAN-FUN Gerät"
	CLicht            = "Licht/Lampe"
	CAlarm            = "Alarm-Sensor"
	CButton           = "AVM-Button"
	CHKR              = "Heizkörperregler"
	CEnergieMesser    = "Energie Messgerät"
	CTempSensor       = "Temperatursensor"
	CSteckdose        = "Schaltsteckdose"
	CRepeater         = "AVM DECT Repeater"
	CMikrofon         = "Mikrofon"
	CHanfunUnit       = "HAN-FUN-Units" // Sozusagen das "Kind" des Han-Fun-Geräts
	CSchaltbar        = "an-/ausschaltbares Gerät/Steckdose/Lampe/Aktor"
	CDimmbar          = "Gerät mit einstellbarem Dimm-, Höhen- bzw. Niveau-Level"
	CLampeMitFarbtemp = "Lampe mit einstellbarer Farbe/Farbtemperatur"
	CRollladen        = "Rollladen(Blind) - hoch, runter, stop und level 0% bis 100 %"

	// these keys are ignored when parsing han-fun units because they are already present in the device-struct
	ignoreKeywords = "-functionbitmask,-fwversion,-id,-identifier,-manufacturer,-productname,Name,present,txbusy,etsiunitinfo"

	EvTastendruckKurz = "lang"
	EvTastendruckLang = "kurz"
)

var (
	hanFunUnitTypes = map[string]string{
		"273": "SIMPLE_BUTTON",
		"256": "SIMPLE_ON_OFF_SWITCHABLE",
		"257": "SIMPLE_ON_OFF_SWITCH",
		"262": "AC_OUTLET",
		"263": "AC_OUTLET_SIMPLE_POWER_METERING",
		"264": "SIMPLE_LIGHT",
		"265": "DIMMABLE_LIGHT",
		"266": "DIMMER_SWITCH",
		"277": "COLOR_BULB",
		"278": "DIMMABLE_COLOR_BULB	",
		"281": "BLIND",
		"282": "LAMELLAR",
		"512": "SIMPLE_DETECTOR",
		"513": "DOOR_OPEN_CLOSE_DETECTOR",
		"514": "WINDOW_OPEN_CLOSE_DETECTOR",
		"515": "MOTION_DETECTOR",
		"518": "FLOOD_DETECTOR",
		"519": "GLAS_BREAK_DETECTOR",
		"520": "VIBRATION_DETECTOR",
		"640": "SIREN",
	}

	hanFunInterfacesStr = map[string]string{
		"277":  "KEEP_ALIVE",
		"256":  "ALERT",
		"512":  "ON_OFF",
		"513":  "LEVEL_CTRL",
		"514":  "COLOR_CTRL",
		"516":  "OPEN_CLOSE",
		"517":  "OPEN_CLOSE_CONFIG",
		"772":  "SIMPLE_BUTTON",
		"1024": "SUOTA-Update",
	}

	hanFunInterfaces = map[string]HFInterface{
		"277":  HFKeepAlive{},
		"256":  HFAlert{},
		"512":  HFOnOff{},
		"513":  HFLevelControl{},
		"514":  HFColorControl{},
		"516":  HFOpenClose{},
		"517":  HFOpenCloseConfig{},
		"772":  HFSimpleButton{},
		"1024": HFSuotaUpdate{},
	}
)
