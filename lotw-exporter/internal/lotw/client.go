package lotw

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Client handles interaction with the LoTW website.
type Client struct {
	Username   string
	Password   string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new LoTW client.
func NewClient(username, password string) *Client {
	return &Client{
		Username: username,
		Password: password,
		BaseURL:  "https://lotw.arrl.org/lotwuser/lotwreport.adi",
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second, // Reports can be slow
		},
	}
}

// FetchReport downloads the ADIF report from LoTW.
// Optional 'since' date can be provided to filter (though LoTW API is a bit basic).
// Actually LoTW allows query by qso_query=1 & qso_startdate=YYYY-MM-DD
func (c *Client) FetchReport(since time.Time) (io.ReadCloser, error) {
	// Construct URL params
	// parameters documented/reversed engineered from standard usage
	// login: user
	// password: password
	// qso_query: 1 (Download Report)
	//          If we want everything, we might use default options.
	//          For 'All QSOs', usually we just ask for qso_query=1
	//          Let's include qso_withown=yes to get own call?

	// Standard fetch URL construction:
	v := url.Values{}
	v.Set("login", c.Username)
	v.Set("password", c.Password)
	v.Set("qso_query", "1") // Download report
	v.Set("qso_qsl", "no")  // "no" = Include all QSOs, not just QSLs. "QSL ONLY: YES" happens if omitted/defaulted?

	// If we want new QSOs only:
	if !since.IsZero() {
		// LoTW format: YYYY-MM-DD
		v.Set("qso_startdate", since.Format("2006-01-02"))
	} else {
		// If since is zero, LoTW defaults to "Since Last Download" or "Now", which returns nothing.
		// We must explicitly ask for everything.
		v.Set("qso_startdate", "1900-01-01")
	}

	// Add extra cols usually needed:
	// qso_owncall, qso_callsign, qso_band, qso_mode, qso_qsl, qso_qsldate, country?
	// ADIF dump usually includes standard fields by default.

	reqURL := fmt.Sprintf("%s?%s", c.BaseURL, v.Encode())

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("lotw api returned status: %d", resp.StatusCode)
	}

	return resp.Body, nil
}
