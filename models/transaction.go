package models

//Transaction represents the incoming transaction payload

type Transaction struct {
	ID     string  `json:"id"` // Unique identifier for the transaction
	UserID string  `json:"userId"` //ID of the user performing the transaction
	Amount float64 `json:"amount"` //Monetary amount in the transaction currency
	Currency string `json:"currency"` //Currency of the transaction
	IPAddress string `json:"ipAddress"` //IP address of the user
	Country   string `json:"country"` //Country of the user
	DeviceID  string `json:"deviceId"` //Device ID of the user's device
	AmountAgeDays int     `json:"amountAgeDays"` //Age of the amount in days
	IsNewDevice bool `json:"isNewDevice"` //Indicates if the device is new
}
//each field has a json tag so it can be easily serialized to and from JSON