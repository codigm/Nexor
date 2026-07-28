package iso20022

import (
	"encoding/xml"
	"time"
)

// Pacs008 represents a simplified ISO 20022 Customer Credit Transfer Initiation message.
// Real ISO 20022 messages are deeply nested, but this captures the essential fields for OLYMPUS.
type Document struct {
	XMLName xml.Name `xml:"Document"`
	Xmlns   string   `xml:"xmlns,attr"`
	FIToFICstmrCdtTrf FIToFICstmrCdtTrf `xml:"FIToFICstmrCdtTrf"`
}

type FIToFICstmrCdtTrf struct {
	GrpHdr      GroupHeader            `xml:"GrpHdr"`
	CdtTrfTxInf CreditTransferTxInfo   `xml:"CdtTrfTxInf"`
}

type GroupHeader struct {
	MsgId    string `xml:"MsgId"`
	CreDtTm  string `xml:"CreDtTm"`
	NbOfTxs  int    `xml:"NbOfTxs"`
	SttlmInf SettlementInfo `xml:"SttlmInf"`
}

type SettlementInfo struct {
	SttlmMtd string `xml:"SttlmMtd"` // e.g., "RTGS" or "NEFT"
}

type CreditTransferTxInfo struct {
	PmtId      PaymentId        `xml:"PmtId"`
	IntrBkSttlmAmt Amount       `xml:"IntrBkSttlmAmt"`
	Dbtr       Party            `xml:"Dbtr"`
	DbtrAcct   Account          `xml:"DbtrAcct"`
	Cdtr       Party            `xml:"Cdtr"`
	CdtrAcct   Account          `xml:"CdtrAcct"`
}

type PaymentId struct {
	EndToEndId string `xml:"EndToEndId"` // Transaction ID
	TxId       string `xml:"TxId"`       // Saga ID
}

type Amount struct {
	Ccy   string `xml:"Ccy,attr"` // Currency
	Value string `xml:",chardata"`
}

type Party struct {
	Nm string `xml:"Nm"` // Name
}

type Account struct {
	Id AccountId `xml:"Id"`
}

type AccountId struct {
	Othr OtherId `xml:"Othr"`
}

type OtherId struct {
	Id string `xml:"Id"` // Account Number
}

// GeneratePacs008 creates a new ISO 20022 XML document for a transaction.
func GeneratePacs008(txId, sagaId, senderAcct, receiverAcct, amount, currency, settlementMode string) ([]byte, error) {
	doc := Document{
		Xmlns: "urn:iso:std:iso:20022:tech:xsd:pacs.008.001.08",
		FIToFICstmrCdtTrf: FIToFICstmrCdtTrf{
			GrpHdr: GroupHeader{
				MsgId:   "MSG-" + txId,
				CreDtTm: time.Now().Format(time.RFC3339),
				NbOfTxs: 1,
				SttlmInf: SettlementInfo{
					SttlmMtd: settlementMode,
				},
			},
			CdtTrfTxInf: CreditTransferTxInfo{
				PmtId: PaymentId{
					EndToEndId: txId,
					TxId:       sagaId,
				},
				IntrBkSttlmAmt: Amount{
					Ccy:   currency,
					Value: amount,
				},
				DbtrAcct: Account{Id: AccountId{Othr: OtherId{Id: senderAcct}}},
				CdtrAcct: Account{Id: AccountId{Othr: OtherId{Id: receiverAcct}}},
			},
		},
	}
	return xml.MarshalIndent(doc, "", "  ")
}

// ParsePacs008 parses an ISO 20022 XML byte array into the Document struct.
func ParsePacs008(data []byte) (*Document, error) {
	var doc Document
	err := xml.Unmarshal(data, &doc)
	return &doc, err
}
