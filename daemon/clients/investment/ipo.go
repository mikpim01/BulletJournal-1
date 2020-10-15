package investment

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/singerdmx/BulletJournal/daemon/persistence"
)

type IPOClient struct {
	BaseTemplateClient
	data *IPOData
}

type IPOData struct {
	IPO []IPO `json:"ipos"`
}

// Serializer for IPO response
type IPO struct {
	ID                             string 	`json:"id"`
	Date                           string 	`json:"date"`
	Time                           string 	`json:"time"`
	Ticker                         string 	`json:"ticker"`
	Exchange                       string 	`json:"exchange"`
	Name                           string 	`json:"name"`
	OpenDateVerified               bool   	`json:"open_date_verified"`
	PricingDate                    string 	`json:"pricing_date"`
	Currency                       string 	`json:"currency"`
	PriceMin                       string 	`json:"price_min"`
	PriceMax                       string 	`json:"price_max"`
	DealStatus                     string 	`json:"deal_status"`
	InsiderLockupDays              string 	`json:"insider_lockup_days"`
	InsiderLockupDate              string 	`json:"insider_lockup_date"`
	OfferingValue                  int    	`json:"offering_value"`
	OfferingShares                 int    	`json:"offering_shares"`
	LeadUnderwriters			   string[] `json:"lead_underwriters"`
	SharesOutstanding              int    	`json:"shares_outstanding"`
	UnderwriterQuietExpirationDate string 	`json:"underwriter_quiet_expiration_date"`
	Notes                          string 	`json:"notes"`
	Updated                        int    	`json:"updated"`
}

const layoutISO = "2006-01-02"

func NewIPOClient() (*TemplateClient, error) {
	c := IPOClient{
		BaseTemplateClient: NewBaseTemplateClient(),
	}
	return &TemplateClient{&c}, nil
}

func (c *IPOClient) FetchData() error {
	fmt.Println("fetching IPO")

	yearFrom, monthFrom, dayFrom := time.Now().Date()
	yearTo, monthTo, dayTo := time.Now().AddDate(0, 1, 0).Date()

	dateFrom := dateFormatter(yearFrom, monthFrom, dayFrom)
	dateTo := dateFormatter(yearTo, monthTo, dayTo)

	// Request for IPO info of incoming 30 days
	url := fmt.Sprintf("https://www.benzinga.com/services/webapps/calendar/ipos?tpagesize=500&parameters[date_from]=%+v&parameters[date_to]=%+v&parameters[importance]=0", dateFrom, dateTo)
	resp, err := c.restClient.R().
		Get(url)

	if err != nil {
		return errors.Wrap(err, "IPO client sending request failed!")
	}

	var data IPOData

	if err := json.Unmarshal(resp.Body(), &data); err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unmarshal earnings response failed: %s", string(resp.Body())))
	}
	c.data = &data
	return nil
}

func (c *IPOClient) SendData() (*[]uint64, *[]uint64, error) {
	if c.data == nil {
		return nil, nil, errors.New("Empty IPO data, please fetch data first.")
	}
	created := make([]uint64, 0)
	modified := make([]uint64, 0)
	for i := range c.data.IPO {
		target := c.data.IPO[i]
		availBefore := target.Date
		t, _ := time.Parse(layoutISO, availBefore)
		dueDate := target.Date
		if len(dueDate) > 10 {
			dueDate = dueDate[0:10] // yyyy-MM-dd
		}
		dueTime := target.Time
		if len(dueTime) > 5 {
			dueTime = dueTime[0:5]
		}
		raw, _ := json.Marshal(target)
		item := persistence.SampleTask{
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
			Metadata:        "INVESTMENT_IPO_RECORD",
			Raw:             string(raw),
			Name:            fmt.Sprintf("%v (%v) goes public on %v", target.Name, target.Ticker, dueDate),
			Uid:             "INVESTMENT_IPO_RECORD_" + target.Ticker,
			AvailableBefore: t,
			DueDate:         dueDate,
			DueTime:         dueTime,
			Pending:         true,
			Refreshable:     true,
			TimeZone:        "America/New_York",
		}
		if entityId, newRecord := c.sampleDao.Upsert(&item); newRecord && entityId > 0 {
			created = append(created, entityId)
		} else if !newRecord && entityId > 0 {
			modified = append(modified, entityId)
		}
	}

	return &created, &modified, nil
}
