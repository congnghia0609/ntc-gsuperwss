/**
 *
 * @author nghiatc
 * @since Oct 01, 2020
 */

package nwss

import (
	"log"
	"strings"

	"github.com/congnghia0609/ntc-gconf/nconf"
)

const (
	// NameDPNWSS depthprice
	NameDPNWSS = "depthprice"
	// NameCSNWSS candlesticks
	NameCSNWSS = "candlesticks"
	// NameHTNWSS historytrade
	NameHTNWSS = "historytrade"
	// NameTKNWSS ticker24h
	NameTKNWSS = "ticker24h"
	// NameCRNWSS cerberus
	NameCRNWSS = "cerberus"
)

// MapSymbol Const
var MapSymbol = make(map[string]string)

// TypeTime type time
var TypeTime = map[string]string{
	"1m":  "1m",
	"5m":  "5m",
	"15m": "15m",
	"30m": "30m",
	"1h":  "1h",
	"2h":  "2h",
	"4h":  "4h",
	"6h":  "6h",
	"12h": "12h",
	"1d":  "1d",
	"1w":  "1w",
}

// InitMapSymbol init map symbol
func InitMapSymbol() {
	c := nconf.GetConfig()
	listpair := c.GetString("market.listpair")
	log.Printf("=========== listpair: %s", listpair)

	//var listpair = "ETH_BTC;KNOW_BTC;KNOW_ETH"
	var arrSymbol = strings.Split(listpair, ";")
	// log.Printf("arrSymbol: ", arrSymbol)
	for i := range arrSymbol {
		symbol := arrSymbol[i]
		MapSymbol[symbol] = symbol
	}
	log.Printf("=========== MapSymbol: %v", MapSymbol)
}

// ReloadMapSymbol update map symbol
func ReloadMapSymbol(listpair string) {
	log.Printf("=========== reloadMapSymbol.listpair: %s", listpair)
	if listpair != "" {
		var arrSymbol = strings.Split(listpair, ";")
		// log.Printf("arrSymbol: ", arrSymbol)
		for i := range arrSymbol {
			symbol := arrSymbol[i]
			//// If not exist, add to MapSymbol
			if _, ok := MapSymbol[symbol]; !ok {
				MapSymbol[symbol] = symbol
			}
		}
		log.Printf("=========== reloadMapSymbol.MapSymbol: %v", MapSymbol)
	}
}
