package types

var Duration = map[string]int64{
	"AXA60970348": 30,
	"AXA50593900": 90,
	"AXA58989052": 180,
	"AXA18984329": 365,
	"AXA22499389": 30,
	"AXA77663784": 30,
	"AXA21668423": 90,
	"AXA87415864": 180,
	"AXA46095180": 365,
}

var PolicyType = map[string]string{
	"AXA60970348": "PalmPay AXA Pass Monthly",
	"AXA50593900": "PalmPay AXA Pass Quarterly",
	"AXA58989052": "PalmPay AXA Pass Bi-Annually",

	"AXA77663784": "PalmPay Digital Health Monthly",
	"AXA21668423": "PalmPay Digital Health Quarterly",
	"AXA87415864": "PalmPay Digital Health Bi-Annually",
	"AXA46095180": "PalmPay Digital Health Yearly",
}

type Revenue struct {
	Revid        string `json:"rev_id"`
	Trxnid       string `json:"trnxid"`
	Msisdn       string `json:"msisdn"`
	ProductId    string `json:"productid"`
	Period       int64  `json:"period"`
	Dateadded    string `json:"dateadded"`
	Paylaod      string `json:"payload"`
	TrackNumber  string `json:"tracknumber"`
	TrackNumber2 string `json:"tracknumber2"`
	Expirydate   string `json:"expirydate"`
	Paytype      string `json:"paytype"`
}

type Response struct {
	ReturnedCode   string   `json:"returnedCode"`
	IsSuccessful   bool     `json:"isSuccessful"`
	Message        string   `json:"message"`
	ReturnedObject []Policy `json:"returnedObject"`
}

type Transaction struct {
	EffectiveDate string `json:"effectiveDate"`
	ExpiryDate    string `json:"expiryDate"`
	TrackNumber   string `json:"trackNumber"`
	TransactionID string `json:"transactionID"`
}

// Define the structure of the returnedObject (Policy)
type Policy struct {
	TrackingNumber  string  `json:"trackingNumber"`
	FirstName       string  `json:"firstName"`
	LastName        string  `json:"lastName"`
	BundleName      string  `json:"bundleName"`
	PartnerName     *string `json:"partnerName"` // Nullable fields are pointers
	PartnerCode     string  `json:"partnerCode"`
	PhoneNumber     *string `json:"phoneNumber"`
	BookingType     *string `json:"bookingType"`
	PolicyStatus    string  `json:"policyStatus"`
	DateOfBirth     *string `json:"dateOfBirth"`
	Gender          string  `json:"gender"`
	PolicyStartDate string  `json:"policyStartDate"`
	PolicyEndDate   string  `json:"policyEndDate"`
	EmailAddress    string  `json:"emailAddress"`
	LocalGovernment string  `json:"localGovernment"`
	Address         string  `json:"address"`
}
