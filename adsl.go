// Package adsl allows one to lookup ADSL/ADSL2+ exchange information by address.
package adsl

import (
	"bytes"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/garfunkel/go-mapregexp"
	"net/http"
	"net/url"
	"strconv"
)

// ADSL3ExchangeURL is the URL to the lookup site.
const ADSL2ExchangesURL = "http://www.adsl2exchanges.com.au/addresslookupstart.php"

// EquipmentProvider is a type representing a provider of ADSL equipment.
type EquipmentProvider struct {
	Name      string
	Status    string
	Estimate  string
	Available bool
}

// Info represents ADSL exchange information for an address.
type Info struct {
	Exchange           string
	Zone               int
	Distance           float32
	CableLength        float32
	EstimatedSpeed     int
	NBNAvailable       bool
	EquipmentProviders []EquipmentProvider
}

// Lookup takes an address and returns ADSL exchange information for it.
func Lookup(address string) (info *Info, err error) {
	info = new(Info)

	address = "ijoaijsdfj  ajisdjf"

	client := &http.Client{}
	data := url.Values{"Address": {address}}
	request, err := http.NewRequest("POST", ADSL2ExchangesURL,
		bytes.NewBufferString(data.Encode()))

	if err != nil {
		return
	}

	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	response, err := client.Do(request)

	if err != nil {
		return
	}

	doc, err := goquery.NewDocumentFromResponse(response)

	if err != nil {
		return
	}

	html, err := doc.Html()

	if err != nil {
		return
	}

	if response.Request.URL.Path == "/error.php" {
		err = errors.New("could not locate address")

		return
	}

	regex := mapregexp.MustCompile(`content: "You are (?P<distance>[\d\.]+) m from (?P<exchange>.*?) as the crow flies\.<br>Estimated cable length of (?P<cablelength>[\d\.]+) m\.<br>Estimated speed of (?P<speed>[\d\.]+)<br>Zone (?P<zone>\d+)<br>"`)
	groups := regex.FindStringSubmatchMap(html)

	if groups == nil {
		err = errors.New("could not parse ADSL availability info")

		return
	}

	info.Exchange = groups["exchange"]
	info.Zone, err = strconv.Atoi(groups["zone"])

	if err != nil {
		err = errors.New("could not parse ADSL exchange zone")

		return
	}

	distance, err := strconv.ParseFloat(groups["distance"], 32)

	if err != nil {
		err = errors.New("could not parse ADSL exchange distance")

		return
	}

	info.Distance = float32(distance)

	cableLength, err := strconv.ParseFloat(groups["cablelength"], 32)

	if err != nil {
		err = errors.New("could not parse ADSL exchange cable length")

		return
	}

	info.CableLength = float32(cableLength)
	info.EstimatedSpeed, err = strconv.Atoi(groups["speed"])

	if err != nil {
		err = errors.New("could not parse ADSL exchange speed")

		return
	}

	nbnInfo := doc.Find("#nbnenabled > #sample > tbody > tr > td")

	if len(nbnInfo.Nodes) != 1 {
		err = errors.New("could not parse NBN availability info")

		return
	}

	if nbnInfo.Text() == "YES" {
		info.NBNAvailable = true
	}

	providerInfo := doc.Find("#eproviders > #sample > tbody > tr")

	if len(providerInfo.Nodes) < 2 {
		err = errors.New("could not parse ADSL providers")

		return
	}

	providerInfo.EachWithBreak(func(i int, sel *goquery.Selection) bool {
		if i%2 == 0 {
			return true
		}

		provider := EquipmentProvider{}
		name := sel.Find("td:nth-child(2)")

		if len(name.Nodes) != 1 {
			err = errors.New("could not parse equipment provider name")

			return false
		}

		provider.Name = name.Text()

		status := sel.Find("td:nth-child(3)")

		if len(status.Nodes) != 1 {
			err = errors.New("could not parse equipment provider status")

			return false
		}

		provider.Status = status.Text()

		estimate := sel.Find("td:nth-child(4)")

		if len(estimate.Nodes) != 1 {
			err = errors.New("could not parse equipment provider estimate")

			return false
		}

		provider.Estimate = estimate.Text()

		availability := sel.Find("td:nth-child(5)")

		if len(availability.Nodes) != 1 {
			err = errors.New("could not parse equipment provider availability")

			return false
		}

		if availability.Text() == "Yes" {
			provider.Available = true
		}

		info.EquipmentProviders = append(info.EquipmentProviders, provider)

		return true
	})

	if err != nil {
		return
	}

	return
}
