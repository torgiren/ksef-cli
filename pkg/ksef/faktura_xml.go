package ksef

import "encoding/xml"

// FakturaXMLNamespace is the target namespace of the KSeF FA(3) schema.
// The date in the URL is part of the versioned namespace — a new major version
// of the schema (FA(4)) would have a different URL, requiring struct updates.
// FA(3) was introduced with mandatory KSeF in 2025 and is stable within this version.
const FakturaXMLNamespace = "http://crd.gov.pl/wzor/2025/06/25/13775/"

// FakturaXML is the root element of a KSeF FA(3) e-invoice.
// Schema: https://github.com/CIRFMF/ksef-docs/blob/main/faktury/schemy/FA/schemat_FA(3)_v1-0E.xsd
type FakturaXML struct {
	XMLName   xml.Name      `xml:"Faktura"`
	Naglowek  NaglowekXML   `xml:"Naglowek"`
	Podmiot1  PodmiotXML    `xml:"Podmiot1"`
	Podmiot2  PodmiotXML    `xml:"Podmiot2"`
	Podmiot3  []Podmiot3XML `xml:"Podmiot3"` // podmioty trzecie, opcjonalne
	Fa        FaXML         `xml:"Fa"`
	Stopka    StopkaXML     `xml:"Stopka"`
	Zalacznik *ZalacznikXML `xml:"Zalacznik"` // opcjonalne załączniki z danymi technicznymi
}

// ── Nagłówek ──────────────────────────────────────────────────────────────────

// KodFormularzaXML contains the form code with schema version attributes.
type KodFormularzaXML struct {
	KodSystemowy string `xml:"kodSystemowy,attr"` // np. "FA (3)"
	WersjaSchemy string `xml:"wersjaSchemy,attr"` // np. "1-0E"
	Value        string `xml:",chardata"`         // np. "FA"
}

// NaglowekXML is the invoice header.
type NaglowekXML struct {
	KodFormularza     KodFormularzaXML `xml:"KodFormularza"`
	WariantFormularza string           `xml:"WariantFormularza"` // zawsze "3"
	DataWytworzeniaFa string           `xml:"DataWytworzeniaFa"` // datetime wytworzenia pliku
	SystemInfo        string           `xml:"SystemInfo"`        // opcjonalne; nazwa systemu wystawiającego
}

// ── Podmioty ──────────────────────────────────────────────────────────────────

// AdresXML holds address lines.
type AdresXML struct {
	KodKraju string `xml:"KodKraju"` // kod kraju, np. "PL"
	AdresL1  string `xml:"AdresL1"`  // pierwsza linia adresu
	AdresL2  string `xml:"AdresL2"`  // druga linia adresu (opcjonalna)
	GLN      string `xml:"GLN"`      // Global Location Number (opcjonalne)
}

// DaneKontaktoweXML holds optional contact data for Podmiot2.
type DaneKontaktoweXML struct {
	Email   string `xml:"Email"`
	Telefon string `xml:"Telefon"`
}

// PodmiotXML represents Podmiot1 (sprzedawca) or Podmiot2 (nabywca).
// Pole NIP jest obowiązkowe; inne pola identyfikacyjne (KodUE/NrVatUE/BrakID)
// są alternatywami dla NIP w przypadkach szczególnych — tu uwzględniony wyłącznie NIP.
type PodmiotXML struct {
	PrefiksPodatnika string            `xml:"PrefiksPodatnika"` // opcjonalne; np. "PL"
	NIP              string            `xml:"DaneIdentyfikacyjne>NIP"`
	KodUE            string            `xml:"DaneIdentyfikacyjne>KodUE"`   // alternatywa NIP: kod UE
	NrVatUE          string            `xml:"DaneIdentyfikacyjne>NrVatUE"` // alternatywa NIP: nr VAT UE
	BrakID           string            `xml:"DaneIdentyfikacyjne>BrakID"`  // "1" gdy podmiot bez NIP
	Nazwa            string            `xml:"DaneIdentyfikacyjne>Nazwa"`
	Adres            AdresXML          `xml:"Adres"`
	DaneKontaktowe   DaneKontaktoweXML `xml:"DaneKontaktowe"` // opcjonalne; tylko Podmiot2
	JST              string            `xml:"JST"`            // Jednostka sektora finansów: 1=tak, 2=nie
	GV               string            `xml:"GV"`             // Grupa VAT: 1=tak, 2=nie
}

// Podmiot3XML represents optional third parties (np. faktorant, odbiorca).
type Podmiot3XML struct {
	NIP     string   `xml:"DaneIdentyfikacyjne>NIP"`
	Nazwa   string   `xml:"DaneIdentyfikacyjne>Nazwa"`
	Adres   AdresXML `xml:"Adres"`
	Rola    string   `xml:"Rola"`    // rola podmiotu
	Rola2   string   `xml:"Rola2"`   // dodatkowa rola
	IDNabyw string   `xml:"IDNabyw"` // identyfikator w systemie nabywcy
}

// ── Faktura główna (Fa) ───────────────────────────────────────────────────────

// OkresFaXML holds the billing period when the invoice covers a date range.
type OkresFaXML struct {
	Od string `xml:"P_6_Od"` // data początkowa okresu
	Do string `xml:"P_6_Do"` // data końcowa okresu
}

// AdnotacjeXML holds mandatory annotation flags.
// Values: "1" = tak, "2" = nie (chyba że opis mówi inaczej).
type AdnotacjeXML struct {
	MetodaKasowa       string `xml:"P_16"`  // metoda kasowa: 1=tak, 2=nie
	Samofakturowanie   string `xml:"P_17"`  // samofakturowanie: 1=tak, 2=nie
	OdwrotneObciazenie string `xml:"P_18"`  // odwrotne obciążenie: 1=tak, 2=nie
	PodzielonaPlatnosc string `xml:"P_18A"` // mechanizm podzielonej płatności: 1=tak, 2=nie
	P_23               string `xml:"P_23"`  // procedura uproszczona OSS/IOSS: 1=tak, 2=nie
}

// DaneFaKorygowanejXML is required when RodzajFaktury = "KOR".
type DaneFaKorygowanejXML struct {
	DataWyst            string `xml:"DataWystFaKorygowanej"`
	NrFaKorygowanej     string `xml:"NrFaKorygowanej"`
	NrKSeF              string `xml:"NrKSeF"`              // "1" jeśli faktura korygowana ma nr KSeF
	NrKSeFFaKorygowanej string `xml:"NrKSeFFaKorygowanej"` // nr KSeF faktury korygowanej
}

// DodatkowyOpisXML holds optional key-value metadata at invoice or line level.
type DodatkowyOpisXML struct {
	NrWiersza string `xml:"NrWiersza"` // opcjonalne; numer pozycji, której dotyczy opis
	Klucz     string `xml:"Klucz"`
	Wartosc   string `xml:"Wartosc"`
}

// FaWierszXML represents a single invoice line item.
// Note on P_8B / P_9A: XSD defines P_8B as "ilość" (quantity) and P_9A as
// "cena jednostkowa netto" (unit price). In practice some issuers use them
// in reversed order — verify with P_11 = P_8B * P_9A.
type FaWierszXML struct {
	Nr      string `xml:"NrWierszaFa"`
	DataPoz string `xml:"P_6A"`    // opcjonalne; data dostawy dla tej pozycji
	Nazwa   string `xml:"P_7"`     // nazwa towaru lub usługi
	Jm      string `xml:"P_8A"`    // jednostka miary
	Ilosc   string `xml:"P_8B"`    // ilość (quantity)
	Cena    string `xml:"P_9A"`    // cena jednostkowa netto
	CenaB   string `xml:"P_9B"`    // cena jednostkowa brutto (art. 106e ust. 7); opcjonalne
	Rabat   string `xml:"P_10"`    // opust/obniżka ceny; opcjonalne
	Netto   string `xml:"P_11"`    // wartość sprzedaży netto
	Brutto  string `xml:"P_11A"`   // wartość sprzedaży brutto (art. 106e ust. 7); opcjonalne
	Vat     string `xml:"P_11Vat"` // kwota podatku (art. 106e ust. 10); opcjonalne
	StVAT   string `xml:"P_12"`    // stawka podatku: 23, 8, 5, 0, ZW, NP, OO
	GTU     string `xml:"GTU"`     // oznaczenie GTU (Grupy Towarowo-Usługowe); opcjonalne
}

// TerminPlatnosciXML holds a single payment deadline.
type TerminPlatnosciXML struct {
	Termin string `xml:"Termin"`
}

// RachunekBankowyXML holds bank account details.
type RachunekBankowyXML struct {
	NrRB         string `xml:"NrRB"`
	SWIFT        string `xml:"SWIFT"`        // opcjonalne
	NazwaBanku   string `xml:"NazwaBanku"`   // opcjonalne
	OpisRachunku string `xml:"OpisRachunku"` // opcjonalne
}

// PlatnoscXML holds payment terms and method.
type PlatnoscXML struct {
	TerminPlatnosci TerminPlatnosciXML `xml:"TerminPlatnosci"`
	// FormaPlatnosci: 1=gotówka 2=karta 3=bon 4=czek 5=kredyt 6=przelew 7=mobilna
	FormaPlatnosci  string             `xml:"FormaPlatnosci"`
	RachunekBankowy RachunekBankowyXML `xml:"RachunekBankowy"`
}

// FaXML is the main invoice body element.
type FaXML struct {
	KodWaluty    string     `xml:"KodWaluty"` // waluta, np. "PLN"
	DataWyst     string     `xml:"P_1"`       // data wystawienia faktury
	MiejsceWyst  string     `xml:"P_1M"`      // miejsce wystawienia (opcjonalne)
	NumerFaktury string     `xml:"P_2"`       // numer faktury
	DataDostawy  string     `xml:"P_6"`       // data dostawy/wykonania usługi (wspólna dla wszystkich pozycji); alternatywa dla OkresFa
	OkresFa      OkresFaXML `xml:"OkresFa"`   // okres rozliczeniowy; alternatywa dla P_6

	// Kwoty w podziale na stawki VAT — wartości ujemne oznaczają korektę.
	KwotaNetto23  string `xml:"P_13_1"`  // netto 23% (lub 22%)
	KwotaVAT23    string `xml:"P_14_1"`  // VAT 23% (lub 22%)
	KwotaVAT23W   string `xml:"P_14_1W"` // VAT 23% w złotych gdy faktura w walucie obcej (opcjonalne)
	KwotaNetto8   string `xml:"P_13_2"`  // netto 8% (lub 7%); opcjonalne
	KwotaVAT8     string `xml:"P_14_2"`  // VAT 8% (lub 7%); opcjonalne
	KwotaVAT8W    string `xml:"P_14_2W"` // VAT 8% w złotych gdy faktura w walucie obcej; opcjonalne
	KwotaNetto5   string `xml:"P_13_3"`  // netto 5%; opcjonalne
	KwotaVAT5     string `xml:"P_14_3"`  // VAT 5%; opcjonalne
	KwotaVAT5W    string `xml:"P_14_3W"` // VAT 5% w złotych gdy faktura w walucie obcej; opcjonalne
	KwotaNetto0   string `xml:"P_13_5"`  // netto 0% (krajowe, bez WDT i eksportu); opcjonalne
	KwotaNettoWDT string `xml:"P_13_6"`  // netto 0% WDT; opcjonalne
	KwotaNettoEKS string `xml:"P_13_7"`  // netto 0% eksport; opcjonalne
	KwotaZW       string `xml:"P_13_8"`  // sprzedaż zwolniona; opcjonalne
	KwotaPozaTer  string `xml:"P_13_9"`  // poza terytorium kraju; opcjonalne
	KwotaOdwObci  string `xml:"P_13_11"` // odwrotne obciążenie; opcjonalne
	KwotaBrutto   string `xml:"P_15"`    // łączna kwota należności (brutto)

	Adnotacje          AdnotacjeXML          `xml:"Adnotacje"`
	RodzajFaktury      string                `xml:"RodzajFaktury"`      // VAT, KOR, ZAL, ROZ, UPR, KOR_ZAL, KOR_ROZ
	DaneFaKorygowanej  *DaneFaKorygowanejXML `xml:"DaneFaKorygowanej"`  // wymagane gdy RodzajFaktury=KOR
	OkresFaKorygowanej string                `xml:"OkresFaKorygowanej"` // okres korygowanej faktury (opcjonalne)
	DodatkowyOpis      []DodatkowyOpisXML    `xml:"DodatkowyOpis"`      // dowolne pary klucz-wartość
	Wiersze            []FaWierszXML         `xml:"FaWiersz"`
	Platnosc           PlatnoscXML           `xml:"Platnosc"`
}

// ── Stopka ────────────────────────────────────────────────────────────────────

// InformacjeXML holds a single footer text block.
type InformacjeXML struct {
	StopkaFaktury string `xml:"StopkaFaktury"`
}

// RejestryXML holds company registry data.
type RejestryXML struct {
	PelnaNazwa string `xml:"PelnaNazwa"` // pełna nazwa (opcjonalne)
	KRS        string `xml:"KRS"`        // numer KRS (opcjonalne)
	REGON      string `xml:"REGON"`      // numer REGON (opcjonalne)
}

// StopkaXML is the invoice footer.
type StopkaXML struct {
	Informacje []InformacjeXML `xml:"Informacje"` // może być kilka bloków tekstu
	Rejestry   RejestryXML     `xml:"Rejestry"`
}

// ── Załącznik ─────────────────────────────────────────────────────────────────

// KolXML describes a column in an attachment table.
type KolXML struct {
	Typ  string `xml:"Typ,attr"` // typ: txt, dec, date
	NKom string `xml:"NKom"`     // nagłówek kolumny
}

// WierszXML is a data row in an attachment table.
type WierszXML struct {
	Komorki []string `xml:"WKom"` // wartości kolejnych kolumn
}

// TabelaXML is a structured table in an attachment block.
type TabelaXML struct {
	TMetaDane []TMetaDanaXML `xml:"TMetaDane"` // metadane tabeli (klucz-wartość)
	Opis      string         `xml:"Opis"`      // opis tabeli
	TNaglowek struct {
		Kol []KolXML `xml:"Kol"`
	} `xml:"TNaglowek"`
	Wiersze []WierszXML `xml:"Wiersz"`
}

// TMetaDanaXML holds a single key-value metadata entry for a table.
type TMetaDanaXML struct {
	TKlucz   string `xml:"TKlucz"`
	TWartosc string `xml:"TWartosc"`
}

// MetaDanaXML holds a single key-value metadata entry for an attachment block.
type MetaDanaXML struct {
	ZKlucz   string `xml:"ZKlucz"`
	ZWartosc string `xml:"ZWartosc"`
}

// BlokDanychXML is a data block in an attachment.
type BlokDanychXML struct {
	MetaDane []MetaDanaXML `xml:"MetaDane"` // ogólne metadane bloku
	Tabele   []TabelaXML   `xml:"Tabela"`
}

// ZalacznikXML holds optional structured attachments (e.g. meter readings, usage tables).
type ZalacznikXML struct {
	BlokDanych BlokDanychXML `xml:"BlokDanych"`
}
