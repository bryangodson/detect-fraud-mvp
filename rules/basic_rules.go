package rules

//rules engine

import "github.com/bryangodson/detect-fraud-mvp/models"

//rule results from simple rules

type RuleResult struct {
	Transaction models.Transaction
	IsFraud      bool //true if rule flags transaction
	Reason       string //reson for flag
}
//basic deterministic checks - fast and safe
func BasicRules(tx models.Transaction)[]RuleResult{
	results:=[]RuleResult{}

	//rule 1 : New account making a large transaction
	if tx.Amount>1000 && tx.AmountAgeDays<7 {
		results=append(results,RuleResult{
			Transaction: tx,
			IsFraud:    true,
			Reason:     "New account making a large transaction",
		})
	}


	//rule 2 : Transaction from a new device and medium-large amout
	if tx.IsNewDevice  && tx.Amount>500{
		results=append(results,RuleResult{
			Transaction: tx,
			IsFraud:    true,
			Reason:     "Transaction from a new device and medium-large amount",
		})
	}
	//rule 3 : Transaction from a high risk country
	highRiskCountries:=map[string] bool{
		"RU":true, //Russia
		"CN":true, //China
		"IR":true, //Iran
		"NG":true, //Nigeria
		"KP":true, //North Korea
		"SY":true, //Syria
		"VE":true, //Venezuela
		"US":true, //United States
		"AF":true, //Afghanistan
	}
	if _, found:=highRiskCountries[tx.Country]; found { 
		/*
		_ ignores the variable, found is set to true if tx.country matches a key in highRiskCountries map
		*/
		results=append(results,RuleResult{
			Transaction: tx,
			IsFraud:    true,
			Reason:     "Transaction is from a high risk country",
		})
	}
	//rules are deterministic and can block obvious fraud without ML latency
	return results
}