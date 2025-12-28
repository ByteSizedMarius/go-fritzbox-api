package aha

// Beschreibung inkl. Fehler aus der Doku übernommen ;)
const (
	TempOff = 253 // radiator off (snowflake)
	TempMax = 254 // radiator max

	CHanfun           = "HAN-FUN Gerät"
	CLicht            = "Licht/Lampe"
	CAlarm            = "Alarm-Sensor"
	CButton           = "AVM-ButtonDevice"
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

	EvTastendruckKurz = "kurz"
	EvTastendruckLang = "lang"
)

var (
	buttonTypes = map[string]bool{
		EvTastendruckKurz: false,
		EvTastendruckLang: false,
	}
)

// these keys are ignored when parsing han-fun units because they are already present in the device-struct
const ignoreKeywords = "-functionbitmask,-fwversion,-id,-identifier,-manufacturer,-productname,Name,present,txbusy,etsiunitinfo"

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
		"278": "DIMMABLE_COLOR_BULB",
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

	hanFunInterfaces = map[string]Interface{
		"277":  KeepAlive{},
		"256":  Alert{},
		"512":  OnOff{},
		"513":  LevelControl{},
		"514":  ColorControl{},
		"516":  OpenClose{},
		"517":  OpenCloseConfig{},
		"772":  SimpleButton{},
		"1024": HFSuotaUpdate{},
	}
)
