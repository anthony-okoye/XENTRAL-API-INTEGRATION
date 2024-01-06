package sync

import (
	"bookbox-backend/internal/database"
	"bookbox-backend/internal/model"
	"bookbox-backend/pkg/logger"
	_ "embed"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type Onix struct {
	Products []Product `xml:"product"`
}

type (
	Product struct {

		/*
		   This field is mandatory and non-repeating.

		   Variable-length, alphanumeric, suggested maximum length 32 characters.
		*/
		RecordReference string `xml:"a001"`

		/*
		   Mandatory and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		NotificationType string `xml:"a002"`

		/*
		   Optional and non-repeating; and may occur only when the <NotificationType>
		   element carries the code value 05.

		   Fixed-length, two numeric digits
		*/
		DeletionCode string `xml:"a198"`

		/*
		   Optional and non-repeating; and may occur only when the <NotificationType>
		   element carries the code value 05.

		   Variable-length text, suggested maximum length 100 characters
		*/
		DeletionText string `xml:"a199"`

		/*
		   Optional and non-repeating, independently
		   of the occurrence of any other field.

		   Fixed-length, two numeric digits
		*/
		RecordSourceType string `xml:"a194"`

		/*
		   Optional and non-repeating, but <RecordSourceIdentifier> must also be present
		   if this field is present.

		   Fixed-length, two numeric digits
		*/
		RecordSourceIdentifierType string `xml:"a195"`

		/*
		   Optional and non-repeating, but <RecordSourceIdentifierType>
		   must also be present if this field is present.

		   Defined by the identifier scheme specified in <RecordSourceIdentifierType>
		*/
		RecordSourceIdentifier string `xml:"a196"`

		/*
		   Optional and non-repeating,
		   independently of the occurrence of any other field.

		   Variable-length text, suggested maximum length 100 characters
		*/
		RecordSourceName string `xml:"a197"`

		ProductIdentifier []ProductIdentifier `xml:"productidentifier"`

		/*
		   Optional, and repeatable if the product carries two or more barcodes from different schemes.
		   The absence of this field does NOT mean that a product is not bar-coded.

		   Fixed-length, 2 numeric digits
		*/
		Barcode string `xml:"b246"`

		/*
		   Mandatory and non-repeating.

		   Fixed-length, two letters.
		*/
		ProductForm string `xml:"b012"`

		/*
		   Optional and repeatable.

		   Fixed-length, four characters: one letter followed by three numeric digits
		*/
		ProductFormDetail string `xml:"b333"`

		ProductFormFeature []ProductFormFeature `xml:"productformfeature"`

		/*
		   Optional and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		ProductPackaging string `xml:"b225"`

		/*
		   The field is optional and non-repeating.

		   Variable-length text, suggested maximum length 200 characters.
		*/
		ProductFormDescription string `xml:"b014"`

		/*
		   This field is optional and non-repeating.

		   Variable-length integer, suggested maximum length 4 digits.
		*/
		NumberOfPieces string `xml:"b210"`

		/*
		   Optional and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		TradeCategory string `xml:"b384"`

		/*
		   Optional and repeatable.

		   Fixed-length, two numeric digits.
		*/
		ProductContentType string `xml:"b385"`

		ContainedItem []ContainedItem `xml:"containeditem"`

		ProductClassification []ProductClassification `xml:"productclassification"`

		/*
		   This element is mandatory if and only if the <ProductForm> code for the product is DG;
		   and non-repeating.

		   Fixed-length, 3 numeric digits
		*/
		EpubType string `xml:"b211"`

		/*
		   Optional and non-repeating, and can occur only if the <EpubType> field is present.

		   Variable-length text, suggested maximum 10 characters
		*/
		EpubTypeVersion string `xml:"b212"`

		/*
		   Optional and non-repeating, and can occur only if the <EpubType> field is present.

		   Variable-length text, suggested maximum 200 characters
		*/
		EpubTypeDescription string `xml:"b213"`

		/*
		   Optional and non-repeating, and can occur only if the <EpubType> field is present.

		   Fixed-length, 2 numeric digits
		*/
		EpubFormat string `xml:"b214"`

		/*
		   Optional and non-repeating, and can occur only if the <EpubFormat> field is present.

		   Variable-length text, suggested maximum 10 characters
		*/
		EpubFormatVersion string `xml:"b215"`

		/*
		   Optional and non-repeating, and can occur only if the <EpubType> field is present,
		   but it does not require the presence of the <EpubFormat> field.

		   Variable-length text, suggested maximum 200 characters
		*/
		EpubFormatDescription string `xml:"b216"`

		/*
		   Optional and non-repeating, and can occur only if the <EpubType> field is present.

		   Fixed-length, 2 numeric digits
		*/
		EpubSource string `xml:"b278"`

		/*
		   Optional and non-repeating, and can occur only if the <EpubSource> field is present.

		   Variable-length text, suggested maximum 10 characters
		*/
		EpubSourceVersion string `xml:"b279"`

		/*
		   Optional and non-repeating, and can occur only if the <EpubType> field is present,
		   but it does not require the presence of the <EpubSource> field.

		   Variable-length text, suggested maximum 200 characters
		*/
		EpubSourceDescription string `xml:"b280"`

		/*
		   Optional and non-repeatable, and can occur only if the <EpubType> field is present.

		   Variable-length text, suggested maximum 200 characters
		*/
		EpubTypeNote string `xml:"b277"`

		Series []Series `xml:"series"`

		/*
		   Optional and non-repeating. Must only be sent in a record that has no instances
		   of the <Series> composite.

		   XML empty element
		*/
		NoSeries string `xml:"n338"`

		Set []Set `xml:"set"`

		Title []Title `xml:"title"`

		WorkIdentifier []WorkIdentifier `xml:"workidentifier"`

		Website []Website `xml:"website"`

		/*
		   Optional and non-repeating.

		   Fixed-length, 2 numeric digits
		*/
		ThesisType string `xml:"b368"`

		/*
		   Optional and non-repeating, but if this element is present, <ThesisType> must also be present.

		   Free text, suggested maximum length 300 characters
		*/
		ThesisPresentedTo string `xml:"b369"`

		/*
		   Optional and non-repeating, but if this element is present, <ThesisType> must also be present.

		   Fixed-length, four numeric digits
		*/
		ThesisYear string `xml:"b370"`

		Contributor []Contributor `xml:"contributor"`

		PersonNameIdentifier PersonNameIdentifier `xml:"personnameidentifier"`

		Name Name `xml:"name"`

		PersonDate PersonDate `xml:"persondate"`

		ProfessionalAffiliation ProfessionalAffiliation `xml:"professionalaffiliation"`

		/*
		   When this field is sent, the receiver should use it to replace all name detail sent in the <Contributor>
		   composite for display purposes only. It does not replace the <BiographicalNote> element. The
		   individual name detail must also be sent in the <Contributor> composite for indexing and retrieval.

		   Variable-length text, suggested maximum length 1000 characters
		*/
		ContributorStatement string `xml:"b049"`

		/*
		   Optional and non-repeating. Must only be sent in a record that has no other elements from Group PR.8.

		   XML empty element
		*/
		NoContributor string `xml:"n339"`

		Conference []Conference `xml:"conference"`

		ConferenceSponsor ConferenceSponsor `xml:"conferencesponsor"`

		ConferenceSponsorIdentifier ConferenceSponsorIdentifier `xml:"conferencesponsoridentifier"`

		/*
		   Optional, and repeatable if the product has characteristics of two or more types (eg revised and annotated).

		   Fixed-length, three upper-case letters
		*/
		EditionTypeCode string `xml:"b056"`

		/*
		   Optional and non-repeating.

		   Variable-length integer, suggested maximum length 4 digits.
		*/
		EditionNumber string `xml:"b057"`

		/*
		   Optional and non-repeating. If this field is used, an <EditionNumber> must also be present.

		   Free form, suggested maximum length 20 characters.
		*/
		EditionVersionNumber string `xml:"b217"`

		/*
		   Optional and non-repeating. When used, the <EditionStatement> must carry a complete description
		   of the nature of the edition, ie it should not be treated as merely supplementary to an
		   <EditionTypeCode> or an <EditionNumber>.

		   Variable-length text, suggested maximum length 100 characters.
		*/
		EditionStatement string `xml:"b058"`

		/*
		   Optional and non-repeating. Must only be sent in a record that has no instances of any of the four
		   preceding Edition elements.

		   XML empty element
		*/
		NoEdition string `xml:"n386"`

		ReligiousText []ReligiousText `xml:"religioustext"`

		Bible Bible `xml:"bible"`

		ReligiousTextFeature ReligiousTextFeature `xml:"religioustextfeature"`

		Language []Language `xml:"language"`

		/*
		   This field is optional, but it is normally required for a printed book unless the <PagesRoman> and
		   <PagesArabic> elements are used, and is non-repeating.

		   Variable length integer, suggested maximum length 6 digits.
		*/
		NumberOfPages string `xml:"b061"`

		/*
		   Optional and non-repeating.

		   Variable length alphabetic, suggested maximum length 10 characters.
		*/
		PagesRoman string `xml:"b254"`

		/*
		   Optional and non-repeating.

		   Variable length numeric, suggested maximum length 6 characters.
		*/
		PagesArabic string `xml:"b255"`

		Extent []Extent `xml:"extent"`

		/*
		   Optional and non-repeating.

		   Variable length integer, suggested maximum length 6 digit
		*/
		NumberOfIllustrations string `xml:"b125"`

		/*
		   The text may also include other content items, eg maps, bibliography, tables, index etc.
		   Optional and non-repeating.

		   Variable length text, suggested maximum length 200 characters.
		*/
		IllustrationsNote string `xml:"b062"`

		Illustrations []Illustrations `xml:"illustrations"`

		/*
		   Optional, and repeatable if the product comprises maps with two or more different scales.

		   Variable length integer, suggested maximum length 6 digits.
		*/
		MapScale string `xml:"b063"`

		/*
		   Optional and non-repeating. Additional BISAC subject category codes may be sent using the <Subject> composite.

		   Fixed-length, three upper-case letters and six numeric digits.
		*/
		BASICMainSubject string `xml:"b064"`

		/*
		   Optional and non-repeating, and may only occur when <BASICMainSubject> is also present.

		   Free form – in practise expected to be an integer or a decimal number such as “2.01”.
		   Suggested maximum length 10 characters, for consistency with other version number elements.
		*/
		BASICVersion string `xml:"b200"`

		/*
		   Optional and non-repeating.

		   Variable-length alphanumeric, suggested maximum length 10 characters to allow for expansion.
		*/
		BICMainSubject string `xml:"b065"`

		/*
		   Optional and non-repeating, and may only occur when <BICMainSubject> is also present.

		   Free form – in practise expected to be an integer. Suggested maximum length 10 characters,
		   for consistency with other version number elements.
		*/
		BICVersion string `xml:"b066"`

		MainSubject []MainSubject `xml:"mainsubject"`

		Subject []Subject `xml:"subject"`

		PersonAsSubject []PersonAsSubject `xml:"personassubject"`

		/*
		   Optional, and repeatable if more than one corporate body is involved.

		   Variable-length text, suggested maximum 200 characters.
		*/
		CorporateBodyAsSubject string `xml:"b071"`

		/*
		   Optional, and repeatable if the subject of the product includes more than one place.

		   Variable-length text, suggested maximum 100 characters.
		*/
		PlaceAsSubject string `xml:"b072"`

		/*
		   Optional, and repeatable if the product is intended for two or more groups.

		   Fixed-length, two numeric digits
		*/
		AudienceCode string `xml:"b073"`

		Audience []Audience `xml:"audience"`

		AudienceRange []AudienceRange `xml:"audiencerange"`

		/*
		   Free text describing the audience for which a product is intended. Optional and non-repeating.

		   Free text, suggested maximum length 1000 characters.
		*/
		AudienceDescription string `xml:"b207"`

		Complexity []Complexity `xml:"complexity"`

		OtherText []OtherText `xml:"othertext"`

		MediaFile []MediaFile `xml:"mediafile"`

		ProductWebsite []ProductWebsite `xml:"productwebsite"`

		Prize []Prize `xml:"prize"`

		ContentItem []ContentItem `xml:"contentitem"`

		TextItem TextItem `xml:"textitem"`

		TextItemIdentifier TextItemIdentifier `xml:"textitemidentifier"`

		PageRun PageRun `xml:"pagerun"`

		Imprint []Imprint `xml:"imprint"`

		Publisher []Publisher `xml:"publisher"`

		/*
		   Optional, and repeatable if the imprint carries two or more cities of publication.

		   Free text, suggested maximum length 50 characters.
		*/
		CityOfPublication string `xml:"b209"`

		/*
		   Optional and non-repeating.

		   Fixed-length, two letters. [Note that ISO 3166-1 specifies that country codes
		   shall be sent as upper case only.]
		*/
		CountryOfPublication string `xml:"b083"`

		/*
		   Optional and non-repeating, but it is very strongly recommended that this element
		   should be included in all ONIX Books Product records, and it is possible that it may
		   be made mandatory in a future release, or that it will be treated as mandatory in national
		   ONIX accreditation schemes.

		   Fixed-length, two numeric digits.
		*/
		PublishingStatus string `xml:"b394"`

		/*
		   Optional and non-repeating, but must be accompanied by the <PublishingStatus>
		   element.

		   Variable-length text, suggested maximum 300 characters.
		*/
		PublishingStatusNote string `xml:"b395"`

		/*
		   Optional and non-repeating.

		   Date as year, month, day (YYYYMMDD)
		*/
		AnnouncementDate string `xml:"b086"`

		/*
		   Optional and non-repeating.

		   Date as year, month, day (YYYYMMDD)
		*/
		TradeAnnouncementDate string `xml:"b362"`

		/*
		   Optional and non-repeating.

		   Four, six or eight numeric digits (YYYY, YYYYMM, or YYYYMMDD).
		*/
		PublicationDate string `xml:"b003"`

		CopyrightStatement []CopyrightStatement `xml:"copyrightstatement"`

		CopyrightOwner CopyrightOwner `xml:"copyrightowner"`

		/*
		   Optional and non-repeating, and may not occur if the <CopyrightStatement>
		   composite is present.

		   Date as year only (YYYY)
		*/
		CopyrightYear string `xml:"b087"`

		/*
		   Optional and non-repeating.

		   Date as year only (YYYY)
		*/
		YearFirstPublished string `xml:"b088"`

		SalesRights []SalesRights `xml:"salesrights"`

		NotForSale []NotForSale `xml:"notforsale"`

		SalesRestriction []SalesRestriction `xml:"salesrestriction"`

		SalesOutlet SalesOutlet `xml:"salesoutlet"`

		Measure []Measure `xml:"measure"`

		RelatedProduct []RelatedProduct `xml:"relatedproduct"`

		/*
		   Optional and non-repeating.

		   Date as year, month, day (YYYYMMDD)
		*/
		OutOfPrintDate string `xml:"h134"`

		SupplyDetail []SupplyDetail `xml:"supplydetail"`

		NewSupplier NewSupplier `xml:"newsupplier"`

		Stock Stock `xml:"stock"`

		BatchBonus BatchBonus `xml:"batchbonus"`

		DiscountCoded DiscountCoded `xml:"discountcoded"`

		Reissue Reissue `xml:"reissue"`

		MarketRepresentation []MarketRepresentation `xml:"marketrepresentation"`

		AgentIdentifier AgentIdentifier `xml:"agentidentifier"`

		MarketDate MarketDate `xml:"marketdate"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 1,000 characters
		*/
		PromotionCampaign string `xml:"k165"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 300 characters
		*/
		PromotionContact string `xml:"k166"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 200 characters
		*/
		InitialPrintRun string `xml:"k167"`

		/*
		   Optional, and repeatable to give information about successive reprintings.

		   Variable-length text, suggested maximum length 200 characters
		*/
		ReprintDetail string `xml:"k309"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 200 characters
		*/
		CopiesSold string `xml:"k168"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 200 characters
		*/
		BookClubAdoption string `xml:"k169"`

		Header []Header `xml:"header"`

		SenderIdentifier SenderIdentifier `xml:"senderidentifier"`

		AddresseeIdentifier AddresseeIdentifier `xml:"addresseeidentifier"`
	}

	ProductIdentifier struct {

		/*
		   Mandatory in each occurrence of the <ProductIdentifier> composite, and non-repeating.

		   Fixed-length, 2 numeric digits
		*/
		ProductIDType string `xml:"b221"`

		/*
		   Optional and non-repeating.

		   Free text, suggested maximum length 50 characters
		*/
		IDTypeName string `xml:"b233"`

		/*
		   Mandatory in each occurrence of the <ProductIdentifier> composite,
		   and non-repeating.

		   According to the identifier type specified in <ProductIDType>
		*/
		IDValue string `xml:"b244"`
	}

	ProductFormFeature struct {

		/*

		   Mandatory in each occurrence of the composite, and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		ProductFormFeatureType string `xml:"b334"`

		/*
		   Presence or absence of this element depends on the <ProductFormFeatureType>,
		   since some product form features (eg thumb index) do not require an accompanying value,
		   while others (eg text font) require free text in <ProductFormFeatureDescription>. Non-repeating.

		   Dependent on the scheme specified in <ProductFormFeatureType>
		*/
		ProductFormFeatureValue string `xml:"b335"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 100 characters.
		*/
		ProductFormFeatureDescription string `xml:"b336"`
	}

	ContainedItem struct {
		ProductIdentifier []ProductIdentifier `xml:"productidentifier"`

		/*
		   Optional and non-repeating.

		   Fixed-length, two letters.
		*/
		ProductForm string `xml:"b012"`

		/*
		   Optional and repeatable

		   Fixed-length, four characters: one letter followed by three numeric digits
		*/
		ProductFormDetail string `xml:"b333"`

		ProductFormFeature []ProductFormFeature `xml:"productformfeature"`

		/*
		   This field can only occur if the <ContainedItem> composite has a <ProductForm>code.

		   Fixed-length, two numeric digits.
		*/
		ProductPackaging string `xml:"b225"`

		/*
		   Optional and non-repeating. This field can only occur if the <ContainedItem> composite
		   has a <ProductForm> code.

		   Variable-length text, suggested maximum length 200 characters.
		*/
		ProductFormDescription string `xml:"b014"`

		/*
		   Optional and non-repeating. This field can only occur if the <ContainedItem>
		   composite has a <ProductForm> code.

		   Variable-length integer, suggested maximum length 4 digits.
		*/
		NumberOfPieces string `xml:"b210"`

		/*
		   Optional and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		TradeCategory string `xml:"b384"`

		/*
		   Optional and repeatable.

		   Fixed-length, two numeric digits.
		*/
		ProductContentType string `xml:"b385"`

		/*
		   Optional and non-repeating.

		   Variable-length integer, maximum four digits
		*/
		ItemQuantity string `xml:"b015"`
	}

	ProductClassification struct {

		/*
		   Mandatory in any instance of the <ProductClassification> composite, and non-repeating.

		   Fixed-length, 2 numeric digits
		*/
		ProductClassificationType string `xml:"b274"`

		/*
		   Mandatory in any instance of the <ProductClassification>
		   composite, and non-repeating.

		   According to the identifier type specified in <ProductClassificationType>
		*/
		ProductClassificationCode string `xml:"b275"`

		/*
		   Optional and non-repeating. Used when a mixed product (eg book and CD) belongs
		   partly to two or more product classifications.

		   Real decimal number in the range 0 to 100
		*/
		Percent string `xml:"b337"`
	}

	Series struct {

		/*
		   Non-repeating. Either the <TitleOfSeries> element or at least one occurrence of the <Title>
		   composite must occur in each occurrence of the <Series> composite.

		   Variable-length text, suggested maximum length 300 characters
		*/
		TitleOfSeries string `xml:"b018"`

		Title []Title `xml:"title"`

		Contributor []Contributor `xml:"contributor"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 20 characters
		*/
		NumberWithinSeries string `xml:"b019"`

		/*
		   Optional and non-repeating.

		   Either four numeric digits, or four numeric digits followed by hyphen followed
		   by four numeric digits
		*/
		YearOfAnnual string `xml:"b020"`
	}

	SeriesIdentifier struct {

		/*
		   Mandatory in each occurrence of the <SeriesIdentifier> composite, and non-repeating.

		   Fixed-length, two numeric digits
		*/
		SeriesIDType string `xml:"b273"`

		/*
		   Optional and non-repeating.

		   Free text, suggested maximum length 50 characters
		*/
		IDTypeName string `xml:"b233"`

		/*
		   Mandatory in each occurrence of the <SeriesIdentifier> composite, and non-repeating.

		   According to the identifier type specified in field PR.5.3
		*/
		IDValue string `xml:"b244"`
	}

	Set struct {
		ProductIdentifier []ProductIdentifier `xml:"productidentifier"`

		/*
		   Non-repeating. Either the <TitleOfSet>element or at least one occurrence of the <Title>
		   composite must occur in each occurrence of the <Set> composite.

		   Variable-length text, suggested maximum length 300 characters
		*/
		TitleOfSet string `xml:"b023"`

		Title []Title `xml:"title"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 20 characters
		*/
		SetPartNumber string `xml:"b024"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 300 characters
		*/
		SetPartTitle string `xml:"b025"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 20 characters
		*/
		ItemNumberWithinSet string `xml:"b026"`

		/*
		   Optional and non-repeating.

		   Variable-length string of integers, each successive integer being separated by
		   a full stop, suggested maximum length 100 characters
		*/
		LevelSequenceNumber string `xml:"b284"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 300 characters
		*/
		SetItemTitle string `xml:"b281"`
	}

	Title struct {

		/*

		   Mandatory in each occurrence of the <Title> composite, and non-repeating.

		   Fixed-length, two numeric digits
		*/
		TitleType string `xml:"b202"`

		/*
		   Optional and non-repeating.

		   Variable-length integer, suggested maximum 3 digits
		*/
		AbbreviatedLength string `xml:"b276"`

		/*
		   Optional and non-repeating: see text at the head of the <Title> composite
		   for details of valid title text options.

		   Variable-length text, suggested maximum 300 characters
		*/
		TitleText string `xml:"b203"`

		/*
		   Optional and non-repeating; can only be used if the <TitleWithoutPrefix> element is also present.

		   Variable-length text, suggested maximum length 20 characters
		*/
		TitlePrefix string `xml:"b030"`

		/*
		   Optional and non-repeating; can only be used if the <TitlePrefix> element is also
		   present.

		   Variable-length text, suggested maximum length 300 characters
		*/
		TitleWithoutPrefix string `xml:"b031"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum 300 characters
		*/
		Subtitle string `xml:"b029"`
	}

	WorkIdentifier struct {

		/*
		   Mandatory in each occurrence of the <WorkIdentifier> composite, and non-repeating.

		   Fixed-length, 2 numeric digits
		*/
		WorkIDType string `xml:"b201"`

		/*
		   Optional and non-repeating.

		   Free text, suggested maximum length 50 characters
		*/
		IDTypeName string `xml:"b233"`

		/*
		   Mandatory in each occurrence of the <WorkIdentifier> composite, and non-repeating.

		   According to the identifier type specified in <WorkIDType>
		*/
		IDValue string `xml:"b244"`
	}

	Website struct {

		/*
		   Optional and non-repeating.

		   Fixed-length, two numeric digits
		*/
		WebsiteRole string `xml:"b367"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 300 characters
		   (XHTML is enabled in this element – see ONIX for Books – Product Information Message
		   – XML Message Specification, Section 7)
		*/
		WebsiteDescription string `xml:"b294"`

		/*
		   Mandatory in each occurrence of the <Website> composite, and non-repeating.

		   Variable-length text, suggested maximum length 300 characters
		*/
		WebsiteLink string `xml:"b295"`
	}

	Contributor struct {

		/*      Optional and non-repeating.

		        Variable-length integer, 1, 2, 3 etc, suggested maximum length 3 digits
		*/
		SequenceNumber string `xml:"b034"`

		/*
		   Mandatory in each occurrence of a <Contributor> composite, and may be repeated if the
		   same person or corporate body has more than one role in relation to the product.

		   Fixed-length, one letter and two numeric digits
		*/
		ContributorRole string `xml:"b035"`

		/*
		   Optional and repeatable in the unlikely event that a single person has been responsible for
		   translation from two or more languages.

		   Fixed-length, three lower-case letters. Note that ISO 639 specifies that these
		   codes should always be in lower-case.
		*/
		LanguageCode string `xml:"b252"`

		/*
		   Optional and non-repeating.

		   Variable-length integer, 1, 2, 3 etc, suggested maximum length 3 digits
		*/
		SequenceNumberWithinRole string `xml:"b340"`

		/*
		   Optional and non-repeating: see Group PR.8 introductory text for valid options.

		   Variable-length text, suggested maximum length 100 characters
		*/
		PersonName string `xml:"b036"`

		/*
		   Optional and non-repeating: see Group PR.8 introductory text for valid options.

		   Variable-length text, suggested maximum length 100 characters
		*/
		PersonNameInverted string `xml:"b037"`

		/*
		   Optional and non-repeating: see Group PR.8 introductory text for valid options.

		   Variable-length text, suggested maximum length 100 characters
		*/
		TitlesBeforeNames string `xml:"b038"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 100 characters
		*/
		NamesBeforeKey string `xml:"b039"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 100 characters
		*/
		PrefixToKey string `xml:"b247"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 100 characters
		*/
		KeyNames string `xml:"b040"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 100 characters

		*/
		NamesAfterKey string `xml:"b041"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 100 characters
		*/
		SuffixToKey string `xml:"b248"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 100 characters
		*/
		LettersAfterNames string `xml:"b042"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 100 characters
		*/
		TitlesAfterNames string `xml:"b043"`

		PersonNameIdentifier PersonNameIdentifier `xml:"personnameidentifier"`

		Name Name `xml:"name"`

		PersonDate PersonDate `xml:"persondate"`

		ProfessionalAffiliation ProfessionalAffiliation `xml:"professionalaffiliation"`

		/*
		   Optional and non-repeating: see Group PR.8 introductory text for valid options.

		   Variable-length text, suggested maximum length 200 characters
		*/
		CorporateName string `xml:"b047"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 500 characters (XHTML is
		   enabled in this element – see ONIX for Books – Product Information Message
		   – XML Message Specification, Section 7)
		*/
		BiographicalNote string `xml:"b044"`

		Website []Website `xml:"website"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 200 characters
		*/
		ContributorDescription string `xml:"b048"`

		/*
		   Optional and non-repeating: see Group PR.8 introductory text for valid options.

		   Fixed-length, two numeric digits
		*/
		UnnamedPersons string `xml:"b249"`

		/*
		   Optional and repeatable.

		   Fixed-length, two letters. [Note that ISO 3166-1 specifies that country codes
		   shall be sent as upper case only.]
		*/
		CountryCode string `xml:"b251"`

		/*
		   Optional and repeatable.

		   Variable-length code, consisting of upper case letters with or without a hyphen,
		   successive codes being separated by spaces. Suggested maximum length 8 characters.
		*/
		RegionCode string `xml:"b398"`
	}
	PersonNameIdentifier struct {

		/*
		   Mandatory in each occurrence of the <PersonNameIdentifier> composite, and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		PersonNameIDType string `xml:"b390"`

		/*
		   Optional and non-repeating.

		   Free text, suggested maximum length 50 character
		*/
		IDTypeName string `xml:"b233"`

		/*
		   Mandatory in each occurrence of the composite, and non-repeating.

		   Determined by the scheme specified in <PersonNameIDType>
		*/
		IDValue string `xml:"b244"`
	}

	Name struct {

		/*
		   Mandatory in each occurrence of the composite, and non-repeating.

		   Fixed-length, two numeric digits
		*/
		PersonNameType string `xml:"b250"`

		/*
		   Within the <Name> composite, all of fields PR.8.5 to PR.8.17 may be used in exactly the same way
		   as specified on preceding pages.
		*/
	}

	PersonDate struct {

		/*
		   Mandatory in each occurrence of the <PersonDate> composite.

		   Fixed-length, three numeric digits
		*/
		PersonDateRole string `xml:"b305"`

		/*
		   Optional and non-repeating. When omitted, the format is assumed to be YYYYMMDD

		   Fixed-length, two numeric digits
		*/
		DateFormat string `xml:"j260"`

		/*
		   Mandatory in each occurrence of the <PersonDate> composite.

		   As specified by the value in <DateFormat>: default YYYYMMDD
		*/
		Date string `xml:"b306"`
	}

	ProfessionalAffiliation struct {

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 100 characters
		*/
		ProfessionalPosition string `xml:"b045"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 100 characters
		*/
		Affiliation string `xml:"b046"`
	}

	Conference struct {

		/*
		   Optional and non-repeating.

		   Fixed-length, two numeric digits
		*/
		ConferenceRole string `xml:"b051"`

		/*
		   This element is mandatory in each occurrence of the <Conference> composite, and non-repeating.

		   Variable-length text, suggested maximum length 200 characters.
		*/
		ConferenceName string `xml:"b052"`

		/*
		   An acronym used as a short form of the name of a conference or conference series given in the
		   <ConferenceName> element. Optional and non-repeating

		   Variable-length text, suggested maximum length 20 characters
		*/
		ConferenceAcronym string `xml:"b341"`

		/*
		   Optional and non-repeating.

		   Variable-length integer, suggested maximum length 4 characters
		*/
		ConferenceNumber string `xml:"b053"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 200 characters
		*/
		ConferenceTheme string `xml:"b342"`

		/*
		   Optional and non-repeating.

		   Date as year (YYYY) or month and year (YYYYMM).
		*/
		ConferenceDate string `xml:"b054"`

		/*
		   The place of a conference to which the product is related. Optional and non-repeating.

		   Variable-length text, suggested maximum length 100 characters
		*/
		ConferencePlace string `xml:"b055"`

		ConferenceSponsor ConferenceSponsor `xml:"conferencesponsor"`

		ConferenceSponsorIdentifier ConferenceSponsorIdentifier `xml:"conferencesponsoridentifier"`

		Website []Website `xml:"website"`
	}

	ConferenceSponsor struct {

		/*
		   The name of a person, used here for a personal sponsor of a conference.

		   Variable-length text, suggested maximum length 100 characters
		*/
		PersonName string `xml:"b036"`

		/*
		   The name of a corporate body, used here for a corporate sponsor of a conference.

		   Variable-length text, suggested maximum length 200 characters
		*/
		CorporateName string `xml:"b047"`
	}

	ConferenceSponsorIdentifier struct {

		/*
		   Mandatory in each occurrence of the <ConferenceSponsorIdentifier> composite, and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		ConferenceSponsorIDType string `xml:"b391"`

		/*
		   Optional and non-repeating.

		   Free text, suggested maximum length 50 characters
		*/
		IDTypeName string `xml:"b233"`

		/*
		   Mandatory in each occurrence of the composite, and non-repeating.

		   Determined by the scheme specified in <ConferenceSponsorIDType>
		*/
		IDValue string `xml:"b244"`
	}

	ReligiousText struct {
		Bible Bible `xml:"bible"`

		/*
		   Mandatory in each occurrence of the <ReligiousText> composite that does not include a
		   <Bible> composite, and non-repeating.

		   Fixed-length, two numeric digits
		*/
		ReligiousTextID string `xml:"b376"`

		ReligiousTextFeature ReligiousTextFeature `xml:"religioustextfeature"`
	}

	Bible struct {

		/*
		   Mandatory in each occurrence of the <Bible> composite, and repeatable so that a list such as Old Testament
		   and Apocrypha can be expressed.

		   Fixed-length, two letters
		*/
		BibleContents string `xml:"b352"`

		/*
		   Mandatory in each occurrence of the <Bible> composite, and repeatable if a work includes text
		   in two or more versions.

		   Fixed-length, three letters
		*/
		BibleVersion string `xml:"b353"`

		/*
		   Optional and non-repeating.

		   Fixed-length, three letters
		*/
		StudyBibleType string `xml:"b389"`

		/*
		   Optional and repeatable.

		   Fixed-length, two letters
		*/
		BiblePurpose string `xml:"b354"`

		/*
		   Optional and non-repeating.

		   Fixed-length, three letters
		*/
		BibleTextOrganization string `xml:"b355"`

		/*
		   Optional and non-repeating.

		   Fixed-length, three letters
		*/
		BibleReferenceLocation string `xml:"b356"`

		/*
		   Optional and repeatable.

		   Fixed-length, two letters
		*/
		BibleTextFeature string `xml:"b357"`
	}

	ReligiousTextFeature struct {

		/*
		   Mandatory in each occurrence of the <ReligiousTextFeature> composite, and non-repeating.

		   Fixed-length, to be confirmed
		*/
		ReligiousTextFeatureType string `xml:"b358"`

		/*
		   Mandatory in each occurrence of the <ReligiousTextFeature> composite, and non-repeating.

		   Fixed-length, to be confirmed
		*/
		ReligiousTextFeatureCode string `xml:"b359"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum 100 characters
		*/
		ReligiousTextFeatureDescription string `xml:"b360"`
	}

	Language struct {

		/*
		   Mandatory in each occurrence of the <Language> composite, and non-repeating.

		   Fixed-length, two numeric digits
		*/
		LanguageRole string `xml:"b253"`

		/*
		   Mandatory in each occurrence of the <Language> composite, and non-repeating.

		   Fixed-length, three lower-case letters. Note that ISO 639 specifies that these
		   codes should always be in lower-case.
		*/
		LanguageCode string `xml:"b252"`

		/*
		   Optional and non-repeating

		   Fixed-length, two letters. [Note that ISO 3166-1 specifies that country codes
		   shall be sent as upper case only.]
		*/
		CountryCode string `xml:"b251"`
	}

	Extent struct {

		/*
		   Mandatory in each occurrence of the <Extent> composite, and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		ExtentType string `xml:"b218"`

		/*
		   Mandatory in each occurrence of the <Extent> composite, and non-repeating.

		   Numeric, with decimal point where required, as specified in field PR.12.4
		*/
		ExtentValue string `xml:"b219"`

		/*
		   Mandatory in each occurrence of the <Extent> composite, and non-repeating.

		   Fixed-length, two numeric digits
		*/
		ExtentUnit string `xml:"b220"`
	}

	Illustrations struct {

		/*
		   Mandatory in each occurrence of the <Illustrations> composite, and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		IllustrationType string `xml:"b256"`

		/*
		   Optional and non-repeating. Required when <IllustrationType>carries the value 00.

		   Variable-length text, suggested maximum length 100 characters.
		*/
		IllustrationTypeDescription string `xml:"b361"`

		/*
		   Optional and non-repeating.

		   Variable-length integer, suggested maximum length 6 digits.
		*/
		Number string `xml:"b257"`
	}

	MainSubject struct {

		/*
		   Mandatory in each occurrence of the composite, and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		MainSubjectSchemeIdentifier string `xml:"b191"`

		/*
		   Optional and non-repeating.

		   Free form. Suggested maximum length 10 characters, for consistency with
		   other version number elements.
		*/
		SubjectSchemeVersion string `xml:"b068"`

		/*
		   Either <SubjectCode> or <SubjectHeadingText> or both must be present in each
		   occurrence of the <MainSubject> composite. Non-repeating.

		   Variable-length, alphanumeric, suggested maximum length 20 characters.
		*/
		SubjectCode string `xml:"b069"`

		/*
		   Either <SubjectCode> or <SubjectHeadingText> or both must be present in each occurrence of the
		   <MainSubject> composite. Non-repeating.

		   Variable-length text, suggested maximum length 100 characters.
		*/
		SubjectHeadingText string `xml:"b070"`
	}

	Subject struct {

		/*
		   Mandatory in each occurrence of the composite, and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		SubjectSchemeIdentifier string `xml:"b067"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 100 characters.
		*/
		SubjectSchemeName string `xml:"b171"`

		/*
		   Optional and non-repeating.

		   Free form. Suggested maximum length 10 characters, for consistency with
		   other version number elements.
		*/
		SubjectSchemeVersion string `xml:"b068"`

		/*
		   Either <SubjectCode> or <SubjectHeadingText> or both must be present in each
		   occurrence of the <Subject> composite. Non-repeating.

		   Variable-length, alphanumeric, suggested maximum length 20 characters.
		*/
		SubjectCode string `xml:"b069"`

		/*
		   Either <SubjectCode> or <SubjectHeadingText> or both must be present in each occurrence
		   of the <Subject> composite. Non-repeating.

		   Variable-length text, suggested maximum length 100 characters.
		*/
		SubjectHeadingText string `xml:"b070"`
	}

	PersonAsSubject struct{}

	Audience struct {

		/*
		   Mandatory in each occurrence of the <Audience> composite, and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		AudienceCodeType string `xml:"b204"`

		/*
		   Optional and non-repeating.

		   Free text, suggested maximum length 50 characters
		*/
		AudienceCodeTypeName string `xml:"b205"`

		/*
		   Mandatory in each occurrence of the <Audience> composite, and non-repeating.

		   Determined by the scheme specified in <AudienceCodeType>.
		*/
		AudienceCodeValue string `xml:"b206"`
	}

	AudienceRange struct {

		/*
		   Mandatory in each occurrence of the <AudienceRange> composite, and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		AudienceRangeQualifier string `xml:"b074"`

		/*
		   Mandatory in each occurrence of the <AudienceRange> composite, and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		AudienceRangePrecision string `xml:"b075"`

		/*
		   A value indicating an exact position within a range, or the upper or lower end of a range.

		   Variable-length string, suggested maximum 10 characters. (This element was
		   previously defined as a variable-length integer, but its definition is extended in
		   ONIX 2.1 to enable certain non-numeric values to be carried. For values that
		   BISAC has defined for US school grades and pre-school levels, see List 77.)
		*/
		AudienceRangeValue string `xml:"b076"`
	}

	Complexity struct {

		/*
		   A n ONIX code specifying the scheme from which the value in <ComplexityCode> is taken.

		   Fixed-length, two numeric digits
		*/
		ComplexitySchemeIdentifier string `xml:"b077"`

		/*
		   A code specifying the level of complexity of a text.

		   Variable-length, alphanumeric, suggested maximum length 20 characters.
		*/
		ComplexityCode string `xml:"b078"`
	}

	OtherText struct {

		/*
		   Mandatory in each occurrence of the <OtherText> composite, and non-repeating.

		   Fixed-length, two characters (initially allocated as 01, 02 etc)
		*/
		TextTypeCode string `xml:"d102"`

		/*
		   Optional and non-repeating.

		   Fixed-length, two numeric digits
		*/
		TextFormat string `xml:"d103"`

		/*
		   Either the <Text> element or both of the <TextLinkType> and <TextLink> elements
		   must be present in any occurrence of the <OtherText> composite. Non-repeating.

		   Variable length text (XHTML is enabled in this element – see ONIX for Books –
		   Product Information Message – XML Message Specification, Section 7)
		*/
		Text string `xml:"d104"`

		/*
		   An ONIX code which identifies the type of link which is given in the <TextLink> element.

		   Fixed-length, two numeric digits
		*/
		TextLinkType string `xml:"d105"`

		/*
		   A link to the text item specified in the <TextTypeCode> element, using the link type specified in
		   <TextLinkType>.

		   Variable-length text, suggested maximum length 300 characters
		*/
		TextLink string `xml:"d106"`

		/*
		   The name of the author of text sent in the <Text> element, or referenced in the <TextLink> element,
		   eg if it is a review or promotional quote.

		   Variable-length text, suggested maximum length 300 characters
		*/
		TextAuthor string `xml:"d107"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 200 characters
		*/
		TextSourceCorporate string `xml:"b374"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 100 characters
		*/
		TextSourceTitle string `xml:"d108"`

		/*
		   The date on which text sent in the <Text> element, or referenced in the <TextLink> element, was
		   published. Optional and non-repeating.

		   Date as four, six or eight digits (YYYY, YYYYMM, YYYYMMDD)
		*/
		TextPublicationDate string `xml:"d109"`

		/*
		   Optional and non-repeating, but either both or neither of <StartDate> and <EndDate> must be present.

		   Fixed-length, 8 numeric digits, YYYYMMDD
		*/
		StartDate string `xml:"b324"`

		/*
		   Optional and non-repeating, but either both or neither of <StartDate> and <EndDate> must be present.

		   Fixed-length, 8 numeric digits, YYYYMMDD
		*/
		EndDate string `xml:"b325"`
	}

	MediaFile struct {

		/*
		   Mandatory in each occurrence of the <MediaFile> composite, and non-repeating.

		   Fixed-length, two characters (initially allocated as 01, 02 etc)
		*/
		MediaFileTypeCode string `xml:"f114"`

		/*
		   For image files, JPEG, GIF and TIF are supported. Optional and non-repeating.

		   Fixed-length, two numeric digits
		*/
		MediaFileFormatCode string `xml:"f115"`

		/*
		   Optional and non-repeating.

		   Variable-length integer, suggested maximum length 6 digits
		*/
		ImageResolution string `xml:"f259"`

		/*
		   Mandatory in each occurrence of the <MediaFile> composite, and non-repeating.

		   Fixed-length, two numeric digits
		*/
		MediaFileLinkTypeCode string `xml:"f116"`

		/*
		   Mandatory in each occurrence of the <MediaFile>composite, and non-repeating.

		   Variable-length text, suggested maximum length 300 characters
		*/
		MediaFileLink string `xml:"f117"`

		/*
		   Optional and non-repeating. Text may include credits, copyright
		   notice, etc. If this field is sent, the individual elements <DownloadCaption>, <DownloadCredit>,
		   and <DownloadCopyrightNotice> must not be sent, and vice versa.

		   Variable-length text, suggested maximum length 1,000 characters (XHTML is
		   enabled in this element – see ONIX for Books – Product Information Message
		   – XML Message Specification, Section 7)
		*/
		TextWithDownload string `xml:"f118"`

		/*
		   Optional and non-repeating. The <DownloadCaption> element may be sent
		   together with either or both of fields <DownloadCredit>, or <DownloadCopyrightNotice>.

		   Variable-length text, suggested maximum length 500 characters (XHTML is
		   enabled in this element – see ONIX for Books – Product Information Message
		   – XML Message Specification, Section 7)
		*/
		DownloadCaption string `xml:"f119"`

		/*
		   Text of a personal or corporate credit associated with a download file, and intended to be displayed
		   whenever the file content is used. Optional and non-repeating.

		   Variable-length text, suggested maximum length 300 characters (XHTML is
		   enabled in this element – see ONIX for Books – Product Information Message
		   – XML Message Specification, Section 7)
		*/
		DownloadCredit string `xml:"f120"`

		/*
		   Text of a copyright notice associated with a download file, and intended to be displayed whenever
		   the file content is used. Optional and non-repeating.

		   Variable-length text, suggested maximum length 300 characters (XHTML is
		   enabled in this element – see ONIX for Books – Product Information Message
		   – XML Message Specification, Section 7)
		*/
		DownloadCopyrightNotice string `xml:"f121"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 500 characters (XHTML is
		   enabled in this element – see ONIX for Books – Product Information Message
		   – XML Message Specification, Section 7)
		*/
		DownloadTerms string `xml:"f122"`

		/*
		   Optional and non-repeating.

		   Fixed-length, 8 numeric digits, YYYYMMDD
		*/
		MediaFileDate string `xml:"f373"`
	}

	ProductWebsite struct {

		/*
		   Optional and non-repeating.

		   Fixed-length, two numeric digits
		*/
		WebsiteRole string `xml:"b367"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 300 characters (XHTML is
		   enabled in this element – see ONIX for Books – Product Information Message
		   – XML Message Specification, Section 7)
		*/
		ProductWebsiteDescription string `xml:"f170"`

		/*
		   Mandatory in each occurrence of the <ProductWebsite>composite, and non-repeating.

		   Variable-length text, suggested maximum length 300 characters
		*/
		ProductWebsiteLink string `xml:"f123"`
	}

	Prize struct {

		/*
		   Mandatory in each occurrence of the <Prize> composite, and non-repeating.

		   Variable-length text, suggested maximum length 100 characters
		*/
		PrizeName string `xml:"g126"`

		/*
		   The year in which a prize or award was given. Optional and non-repeating.

		   Four digits, YYYY
		*/
		PrizeYear string `xml:"g127"`

		/*
		   Optional and non-repeating.

		   Fixed-length, two letters. [Note that ISO 3166-1 specifies that country codes
		   shall be sent as upper case only.]
		*/
		PrizeCountry string `xml:"g128"`

		/*
		   Optional and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		PrizeCode string `xml:"g129"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 500 characters (XHTML is
		   enabled in this element – see ONIX for Books – Product Information Message
		   – XML Message Specification, Section 7)
		*/
		PrizeJury string `xml:"g343"`
	}

	ContentItem struct {

		/*
		   A number which specifies the position of a content item in a multi-level hierarchy of such items.

		   Variable-length string of integers, each successive integer being separated by
		   a full stop, suggested maximum length 100 characters
		*/
		LevelSequenceNumber string `xml:"b284"`

		TextItem TextItem `xml:"textitem"`

		TextItemIdentifier TextItemIdentifier `xml:"textitemidentifier"`

		PageRun PageRun `xml:"pagerun"`

		Website []Website `xml:"website"`

		/*
		   Optional and non-repeating; but either this field or a title (in <DistinctiveTitle> or in a <Title>composite)
		   or both must be present in any occurrence of the <ContentItem> composite.

		   Variable-length alphanumeric, suggested maximum length 20 characters
		*/
		ComponentTypeName string `xml:"b288"`

		/*
		   Optional and non-repeating.

		   Variable-length alphanumeric, suggested maximum length 20 characters
		*/
		ComponentNumber string `xml:"b289"`

		Title []Title `xml:"title"`

		WorkIdentifier []WorkIdentifier `xml:"workidentifier"`

		Contributor []Contributor `xml:"contributor"`

		/*
		   It does not replace any biographical notes sent in the composite. The individual contributor
		   elements must also be sent for indexing and retrieval.

		   Variable-length text, suggested maximum length 1000 characters
		*/
		ContributorStatement string `xml:"b049"`

		Subject []Subject `xml:"subject"`

		PersonAsSubject []PersonAsSubject `xml:"personassubject"`

		/*
		   Optional, and repeatable if more than one corporate body is involved.

		   Variable-length text, suggested maximum 200 characters.
		*/
		CorporateBodyAsSubject string `xml:"b071"`

		/*
		   Optional, and repeatable if the subject of the content item includes more than one place.

		   Variable-length text, suggested maximum 100 characters.
		*/
		PlaceAsSubject string `xml:"b072"`

		OtherText []OtherText `xml:"othertext"`

		MediaFile []MediaFile `xml:"mediafile"`
	}

	TextItem struct {

		/*
		   Mandatory in each occurrence of the <TextItem> composite, and non-repeatable.

		   Fixed length, 2 numeric digits
		*/
		TextItemType string `xml:"b290"`

		TextItemIdentifier TextItemIdentifier `xml:"textitemidentifier"`

		/*
		   Optional and non-repeating; required when the text item is being referenced as part of a structured table of contents.

		   Variable-length alphanumeric, suggested maximum length 20 characters
		*/
		FirstPageNumber string `xml:"b286"`

		/*
		   Optional and non-repeating, and can occur only when <FirstPageNumber> is also present.

		   Variable-length alphanumeric, suggested maximum length 20 characters
		*/
		LastPageNumber string `xml:"b287"`

		PageRun PageRun `xml:"pagerun"`

		/*
		   Optional and non-repeating, but normally expected when the text item is being referenced
		   as part of a structured table of contents.

		   Variable length integer, suggested maximum length 6 digits.
		*/
		NumberOfPages string `xml:"b061"`
	}

	TextItemIdentifier struct {

		/*
		   Mandatory in each occurrence of the <TextItemIdentifier> composite, and non-repeating.

		   Fixed-length, 2 numeric digits
		*/
		TextItemIDType string `xml:"b285"`

		/*
		   Optional and non-repeating.

		   Free text, suggested maximum length 50 character
		*/
		IDTypeName string `xml:"b233"`

		/*
		   Mandatory in each occurrence of the <TextItemIdentifier> composite, and non-repeating.

		   According to the identifier type specified in <TextItemIDType>
		*/
		IDValue string `xml:"b244"`
	}

	PageRun struct {

		/*
		   Mandatory in each occurrence of the <PageRun> composite, and non-repeating.

		   Variable-length alphanumeric, suggested maximum length 20 characters
		*/
		FirstPageNumber string `xml:"b286"`

		/*
		   This element is omitted if an item begins and ends on the same page;
		   otherwise it should occur once and only once in each occurrence of the <PageRun> composite.

		   Variable-length alphanumeric, suggested maximum length 20 characters
		*/
		LastPageNumber string `xml:"b287"`
	}

	Imprint struct {

		/*
		   Optional and non-repeating, but mandatory if the <Imprint> composite does not carry an <ImprintName>.

		   Fixed-length, two numeric digits.
		*/
		NameCodeType string `xml:"b241"`

		/*
		   Optional and non-repeating.

		   Free text, suggested maximum length 50 characters
		*/
		NameCodeTypeName string `xml:"b242"`

		/*
		   Mandatory if and only if <NameCodeType> is present, and non-repeating.

		   Determined by the scheme specified in <NameCodeType>
		*/
		NameCodeValue string `xml:"b243"`

		/*
		   Mandatory if there is no name code in an occurrence of the <Imprint> composite, and optional if a
		   name code is included. Non-repeating.

		   Variable length text, suggested maximum length 100 characters.
		*/
		ImprintName string `xml:"b079"`
	}

	Publisher struct {

		/*
		   Optional and non-repeating. The default if the element is omitted is “publisher”.

		   Fixed-length, two numeric digits.
		*/
		PublishingRole string `xml:"b291"`

		/*
		   Optional and non-repeating, but mandatory if the <Publisher> composite does not carry a
		   <PublisherName>.

		   Fixed-length, two numeric digits.
		*/
		NameCodeType string `xml:"b241"`

		/*
		   Optional and non-repeating.

		   Free text, suggested maximum length 50 characters
		*/
		NameCodeTypeName string `xml:"b242"`

		/*
		   Mandatory if and only if <NameCodeType> is present, and non-repeating.

		   Determined by the scheme specified in <NameCodeType>
		*/
		NameCodeValue string `xml:"b243"`

		/*
		   Mandatory if there is no name code in an occurrence of the <Publisher> composite,
		   and optional if a name code is included. Non-repeating.

		   Variable length text, suggested maximum length 100 characters.
		*/
		PublisherName string `xml:"b081"`

		Website []Website `xml:"website"`
	}

	CopyrightStatement struct {

		/*
		   Mandatory in each occurrence of the <CopyrightStatement> composite, and repeatable if several years are listed.

		   Date as year only (YYYY)
		*/
		CopyrightYear string `xml:"b087"`
	}

	CopyrightOwner struct {
		CopyrightOwnerIdentifier []CopyrightOwnerIdentifier `xml:"copyrightowneridentifier"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 100 characters
		*/
		PersonName string `xml:"b036"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 200 characters
		*/
		CorporateName string `xml:"b047"`
	}

	CopyrightOwnerIdentifier struct {

		/*
		   Mandatory in each occurrence of the <CopyrightOwnerIdentifier> composite, and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		CopyrightOwnerIDType string `xml:"b392"`

		/*
		   Optional and non-repeating.

		   Free text, suggested maximum length 50 characters
		*/
		IDTypeName string `xml:"b233"`

		/*
		   Mandatory in each occurrence of the <CopyrightOwnerIdentifier> composite, and non-repeating.

		   Determined by the scheme specified in <CopyrightOwnerIDType>
		*/
		IDValue string `xml:"b244"`
	}

	SalesRights struct {

		/*
		   Mandatory in each occurrence of the <SalesRights>composite, and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		SalesRightsType string `xml:"b089"`

		/*
		   At least one occurrence of <RightsCountry> or <RightsTerritory> or <RightsRegion> is mandatory
		   in any occurrence of the<SalesRights> composite.

		   One or more fixed-length codes, each with two upper case letters, successive
		   codes being separated by spaces. Suggested maximum length 600
		   characters. [Note that ISO 3166-1 specifies that country codes shall be sent
		   as upper case only.]
		*/
		RightsCountry string `xml:"b090"`

		/*
		   Optional and non-repeating.

		   One or more variable-length codes, each consisting of upper case letters with
		   or without a hyphen, successive codes being separated by spaces.
		   Suggested maximum length 100 characters.
		*/
		RightsTerritory string `xml:"b388"`
	}

	NotForSale struct {

		/*
		   At least one occurrence of <RightsCountry> or <RightsTerritory> is
		   mandatory in each occurrence of the<NotForSale> composite.

		   One or more fixed-length codes, each with two upper case letters, successive
		   codes being separated by spaces. Suggested maximum length 600
		   characters. [Note that ISO 3166-1 specifies that country codes shall be sent
		   as upper case only.]

		*/
		RightsCountry string `xml:"b090"`

		/*
		   Optional and non-repeating.

		   One or more variable-length codes, each consisting of upper case letters with
		   or without a hyphen, successive codes being separated by spaces.
		   Suggested maximum length 100 characters.
		*/
		RightsTerritory string `xml:"b388"`

		ProductIdentifier []ProductIdentifier `xml:"productidentifier"`

		/*
		   Optional and non-repeating.

		   Variable length text, suggested maximum length 100 characters.
		*/
		PublisherName string `xml:"b081"`
	}

	SalesRestriction struct {

		/*
		   Mandatory in each occurrence of the <SalesRestriction> composite, and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		SalesRestrictionType string `xml:"b381"`

		SalesOutlet SalesOutlet `xml:"salesoutlet"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 300 characters
		*/
		SalesRestrictionDetail string `xml:"b383"`
	}

	SalesOutlet struct {
		SalesOutletIdentifier SalesOutletIdentifier `xml:"salesoutletidentifier"`

		/*
		   The name of a wholesale or retail sales outlet to which a sales restriction is linked. Non-repeating.

		   Variable-length text, suggested maximum length 200 characters
		*/
		SalesOutletName string `xml:"b382"`
	}

	SalesOutletIdentifier struct {

		/*
		   Mandatory in each occurrence of the <SalesOutletIdentifier> composite, and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		SalesOutletIDType string `xml:"b393"`

		/*
		   Optional and non-repeating.

		   Free text, suggested maximum length 50 characters
		*/
		IDTypeName string `xml:"b233"`

		/*
		   Mandatory in each occurrence of the <SalesOutletIdentifier> composite, and non-repeating.

		   Determined by the scheme specified in <SalesOutletIDType>
		*/
		IDValue string `xml:"b244"`
	}

	Measure struct {

		/*
		   Mandatory in each occurrence of the <Measure> composite, and non-repeating.

		   Fixed-length, two numeric digits
		*/
		MeasureTypeCode string `xml:"c093"`

		/*
		   Mandatory in each occurrence of the <Measure>composite, and non-repeating.

		   Variable length real number, with an explicit decimal point when required,
		   suggested maximum length 6 characters including a decimal point.
		*/
		Measurement string `xml:"c094"`

		/*
		   Mandatory in each occurrence of the <Measure> composite, and non-repeating.

		   Fixed-length, two letters
		*/
		MeasureUnitCode string `xml:"c095"`
	}

	RelatedProduct struct {

		/*
		   Mandatory in each occurrence of the <RelatedProduct> composite, and non-repeating.

		   Fixed length, two numeric digits
		*/
		RelationCode string `xml:"h208"`

		ProductIdentifier []ProductIdentifier `xml:"productidentifier"`

		Website []Website `xml:"website"`

		/*
		   Optional and non-repeating; required in any occurrence of the <RelatedProduct>
		   composite that does not carry a product identifier.

		   Fixed-length, two letters.
		*/
		ProductForm string `xml:"b012"`

		/*
		   Optional and repeatable.

		   Fixed-length, four characters: one letter followed by three numeric digits
		*/
		ProductFormDetail string `xml:"b333"`

		ProductFormFeature []ProductFormFeature `xml:"productformfeature"`

		/*
		   Optional and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		ProductPackaging string `xml:"b225"`

		/*
		   The field is optional and non-repeating.

		   Variable-length text, suggested maximum length 200 characters.
		*/
		ProductFormDescription string `xml:"b014"`

		/*
		   This field is optional and non-repeating.

		   Variable-length integer, suggested maximum length 4 digits.
		*/
		NumberOfPieces string `xml:"b210"`

		/*
		   Optional and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		TradeCategory string `xml:"b384"`

		/*
		   Optional and repeatable.

		   Fixed-length, two numeric digits.
		*/
		ProductContentType string `xml:"b385"`

		/*
		   This element is mandatory if and only if the <ProductForm> code for the product is DG.

		   Fixed-length, 3 numeric digits
		*/
		EpubType string `xml:"b211"`

		/*
		   Optional and non-repeating, and can occur only if the <EpubType> field is present.

		   Variable-length text, suggested maximum 10 characters
		*/
		EpubTypeVersion string `xml:"b212"`

		/*
		   Optional and non-repeating, and can occur only if the <EpubType> field is present.

		   Variable-length text, suggested maximum 200 characters
		*/
		EpubTypeDescription string `xml:"b213"`

		/*
		   Optional and non-repeating, and can occur only if the <EpubType> field is present.

		   Fixed-length, 2 numeric digits
		*/
		EpubFormat string `xml:"b214"`

		/*
		   Optional and non-repeating, and can occur only if the <EpubFormat> field is present.

		   Variable-length text, suggested maximum 10 characters
		*/
		EpubFormatVersion string `xml:"b215"`

		/*
		   Optional and non-repeating, and can occur only if the <EpubType> field is present,
		   but it does not require the presence of the <EpubFormat> field.

		   Variable-length text, suggested maximum 200 characters
		*/
		EpubFormatDescription string `xml:"b216"`

		/*
		   Optional and non-repeatable, and can occur only if the <EpubType> field is present.

		   Variable-length text, suggested maximum 200 characters
		*/
		EpubTypeNote string `xml:"b277"`

		Publisher []Publisher `xml:"publisher"`
	}

	SupplyDetail struct {
		Price []Price `xml:"price"`

		/*
		   Optional, but each occurrence of the <SupplyDetail> composite must carry either at least one supplier identifier,
		   or a <SupplierName>.

		   Fixed-length, thirteen numeric digits, of which the last is a check digit.
		*/
		SupplierEANLocationNumber string `xml:"j135"`

		/*
		   Optional, but each occurrence of the <SupplyDetail> composite must carry either at least one supplier identifier,
		   or a <SupplierName>.

		   Fixed-length, seven characters. The first six are numeric digits, and the
		   seventh is a check character which may be a numeric digit or letter X.
		*/
		SupplierSAN string `xml:"j136"`

		SupplierIdentifier []SupplierIdentifier `xml:"supplieridentifier"`

		/*
		   Optional and non-repeating; required if no supplier identifier is sent.

		   Variable-length text, suggested maximum length 100 characters
		*/
		SupplierName string `xml:"j137"`

		/*
		   Optional and repeatable.

		   Variable-length text, suggested maximum length 20 characters
		*/
		TelephoneNumber string `xml:"j270"`

		/*
		   Optional and repeatable.

		   Variable-length text, suggested maximum length 20 characters
		*/
		FaxNumber string `xml:"j271"`

		/*
		   Optional and repeatable.

		   Variable-length text, suggested maximum length 100 characters
		*/
		EmailAddress string `xml:"j272"`

		Website []Website `xml:"website"`

		/*
		   Optional and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		SupplierRole string `xml:"j292"`

		/*
		   Optional and repeatable.

		   One or more fixed-length codes, each with two upper case letters, successive
		   codes being separated by spaces. Suggested maximum length 600
		   characters. [Note that ISO 3166-1 specifies that country codes shall be sent
		   as upper case only.]
		*/
		SupplyToCountry string `xml:"j138"`

		/*
		   Optional and non-repeating.

		   One or more variable-length codes, each consisting of upper case letters with
		   or without a hyphen, successive codes being separated by spaces.
		   Suggested maximum length 100 characters.
		*/
		SupplyToTerritory string `xml:"j397"`

		/*
		   Optional and repeatable.

		   One or more fixed-length codes, each with two upper case letters, successive
		   codes being separated by spaces. Suggested maximum length 300
		   characters. [Note that ISO 3166-1 specifies that country codes shall be sent
		   as upper case only.]
		*/
		SupplyToCountryExcluded string `xml:"j140"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 300 characters
		*/
		SupplyRestrictionDetail string `xml:"j399"`

		/*
		   Optional and non-repeating, but this field must be present if <ReturnsCode> is present.

		   Fixed-length, 2 numeric digits
		*/
		ReturnsCodeType string `xml:"j268"`

		/*
		   Optional and non-repeating, but this field must be present if <ReturnsCodeType> is present.

		   According to the scheme specified in <ReturnsCodeType>: for values defined
		   by BISAC for US use, see List 66
		*/
		ReturnsCode string `xml:"j269"`

		/*
		   Optional and non-repeating.

		   Date as year, month, day (YYYYMMDD)
		*/
		LastDateForReturns string `xml:"j387"`

		/*
		   Each occurrence of the <SupplyDetail>
		   composite must carry either <AvailabilityCode> or <ProductAvailability>, or both, but
		   <ProductAvailability> is now strongly preferred. Non-repeating.

		   Fixed-length, two letters
		*/
		AvailabilityCode string `xml:"j141"`

		/*
		   The element is non-repeating. Recommended practise is in future
		   to use this new element, and, where possible and appropriate, to include the <PublishingStatus>
		   element in PR.20.

		   Fixed-length, two numeric digits
		*/
		ProductAvailability string `xml:"j396"`

		NewSupplier NewSupplier `xml:"newsupplier"`

		/*
		   Optional an non-repeating. If the field is omitted, the default format YYYYMMDD will be assumed.

		   Fixed-length, 2 numeric digits
		*/
		DateFormat string `xml:"j260"`

		/*
		   Optional and non-repeating; required with certain code values in the <AvailabilityCode> element.
		   The format is as specified in the <DateFormat> field.

		   Date as year, month, day (YYYYMMDD) or as specified in <DateFormat>
		*/
		ExpectedShipDate string `xml:"j142"`

		/*
		   Optional and non-repeating.

		   Date as year, month, day (YYYYMMDD)
		*/
		OnSaleDate string `xml:"j143"`

		/*
		   Optional and non-repeating.

		   Variable-length integer, one or two digits only
		*/
		OrderTime string `xml:"j144"`

		Stock Stock `xml:"stock"`

		/*
		   (This element is placed in Group PR.24 since it cannot be assumed that pack quantities
		   will be the same for stock held at different suppliers.)

		   Variable-length integer, suggested maximum length four digits
		*/
		PackQuantity string `xml:"j145"`

		/*
		   Optional and non-repeating.

		   Provisional: fixed-length, single letter
		*/
		AudienceRestrictionFlag string `xml:"j146"`

		/*
		   Optional and non-repeating.

		   Variable-length text, maximum 300 characters
		*/
		AudienceRestrictionNote string `xml:"j147"`

		/*
		   Optional and non-repeating, but required if the <SupplyDetail> composite does not carry a price.

		   Fixed-length, two numeric digits.
		*/
		UnpricedItemType string `xml:"j192"`
	}

	SupplierIdentifier struct {

		/*
		   Mandatory in each occurrence of the <SupplierIdentifier> composite, and non-repeating.

		   Fixed-length, 2 numeric digits
		*/
		SupplierIDType string `xml:"j345"`

		/*
		   Optional and non-repeating.

		   Free text, suggested maximum length 50 characters
		*/
		IDTypeName string `xml:"b233"`

		/*
		   Mandatory in each occurrence of the <SupplierIdentifier> composite, and non-repeating.

		   According to the identifier type specified in <SupplierIDType>
		*/
		IDValue string `xml:"b244"`
	}

	NewSupplier struct {

		/*
		   Optional and non-repeating, but each occurrence of the <NewSupplier>
		   composite must carry either at least one supplier identifier, or a <SupplierName>.

		   Fixed-length, thirteen numeric digits, of which the last is a check digit.
		*/
		SupplierEANLocationNumber string `xml:"j135"`

		/*
		   Optional, but each occurrence of the <NewSupplier> composite must carry either at least one supplier
		   identifier, or a <SupplierName>.

		   Fixed-length, seven characters. The first six are numeric digits, and the
		   seventh is a check character which may be a numeric digit or letter X.
		*/
		SupplierSAN string `xml:"j136"`

		SupplierIdentifier []SupplierIdentifier `xml:"supplieridentifier"`

		/*
		   Optional and non-repeating; required if no supplier identifier is sent in an
		   occurrence of the <NewSupplier> composite.

		   Variable-length text, suggested maximum length 100 characters
		*/
		SupplierName string `xml:"j137"`

		/*
		   Optional and repeatable.

		   Variable-length text, suggested maximum length 20 characters
		*/
		TelephoneNumber string `xml:"j270"`

		/*
		   Optional and repeatable.

		   Variable-length text, suggested maximum length 20 characters
		*/
		FaxNumber string `xml:"j271"`

		/*
		   Optional and repeatable.

		   Variable-length text, suggested maximum length 100 characters
		*/
		EmailAddress string `xml:"j272"`
	}

	Stock struct {
		LocationIdentifier LocationIdentifier `xml:"locationidentifier"`

		/*
		   The name of a stock location. Optional and non-repeating.

		   Free text, suggested maximum length 100 characters
		*/
		LocationName string `xml:"j349"`

		StockQuantityCoded StockQuantityCoded `xml:"stockquantitycoded"`

		/*
		   Either <StockQuantityCoded> or <OnHand> is mandatory in each occurrence of the <Stock> composite,
		   even if the onhand quantity is zero. Non-repeating.

		   Variable-length integer, suggested maximum length 7 digits
		*/
		OnHand string `xml:"j350"`

		/*
		   The quantity of stock on order. Optional and non-repeating.

		   Variable-length integer, suggested maximum length 7 digits
		*/
		OnOrder string `xml:"j351"`

		/*
		   Optional and non-repeating.

		   Variable-length integer, suggested maximum length 7 digits
		*/
		CBO string `xml:"j375"`

		OnOrderDetail OnOrderDetail `xml:"onorderdetail"`
	}

	LocationIdentifier struct {

		/*
		   Mandatory in each occurrence of the <LocationIdentifier> composite, and non-repeating.

		   Fixed-length, 2 numeric digits
		*/
		LocationIDType string `xml:"j377"`

		/*
		   Optional and non-repeating.

		   Free text, suggested maximum length 50 characters
		*/
		IDTypeName string `xml:"b233"`

		/*
		   Mandatory in each occurrence of the <LocationIdentifier> composite, and non-repeating.

		   According to the identifier type specified in <LocationIDType>
		*/
		IDValue string `xml:"b244"`
	}

	StockQuantityCoded struct {

		/*
		   Mandatory in each occurrence of the <StockQuantityCoded> composite, and non-repeating.

		   Fixed-length, 2 numeric digits
		*/
		StockQuantityCodeType string `xml:"j293"`

		/*
		   Optional, and non-repeating.

		   Free text, suggested maximum length 50 characters
		*/
		StockQuantityCodeTypeName string `xml:"j296"`

		/*
		   Mandatory in each occurrence of the <StockQuantityCoded> composite, and non-repeating.

		   According to the scheme specified in <StockQuantityCodeType>
		*/
		StockQuantityCode string `xml:"j297"`
	}

	OnOrderDetail struct {

		/*
		   Mandatory in each occurrence of the <OnOrderDetail> composite, and non-repeating.

		   Variable-length integer, suggested maximum length 7 digits
		*/
		OnOrder string `xml:"j351"`

		/*
		   Mandatory in each occurrence of the <OnOrderDetail> composite, and non-repeating.

		   Fixed-length, 8 numeric digits, YYYYMMDD
		*/
		ExpectedDate string `xml:"j302"`
	}

	Price struct {

		/*
		   Optional, provided that a <DefaultPriceTypeCode> has been specified in the message
		   header, and non-repeating.

		   Fixed-length, two numeric digits
		*/
		PriceTypeCode string `xml:"j148"`

		/*
		   Optional and non-repeating.

		   Fixed-length, two numeric digits
		*/
		PriceQualifier string `xml:"j261"`

		/*
		   Optional and non-repeating.

		   Text, suggested maximum length 200 characters
		*/
		PriceTypeDescription string `xml:"j262"`

		/*
		   Optional and non-repeating.

		   Fixed-length, two numeric digits
		*/
		PricePer string `xml:"j239"`

		/*
		   Optional and non-repeating.

		   Variable-length integer, suggested maximum length 4 digits
		*/
		MinimumOrderQuantity string `xml:"j263"`

		BatchBonus BatchBonus `xml:"batchbonus"`

		/*
		   This element should be used only in the absence of a “Default class of trade”
		   <m193> in the message header, or when the class of trade is other than the default.

		   Text, suggested maximum length 50 characters
		*/
		ClassOfTrade string `xml:"j149"`

		/*
		   This code does not identify an absolute rate of discount, but it allows a bookseller to derive the actual discount
		   by reference to a look-up table provided separately by the supplier.

		   Fixed-length, 8 characters
		   Position 1 A (identifying BIC as the source of the supplier code)
		   Positions 2-5 Supplier code, alphabetical, assigned by BIC
		   Positions 6-8 Discount group code, alphanumeric, assigned by the
		   supplier. If less than three characters, the code is left
		   justified and unused positions are sent as spaces.
		*/
		BICDiscountGroupCode string `xml:"j150"`

		DiscountCoded DiscountCoded `xml:"discountcoded"`

		/*
		   Optional and non-repeating.

		   Variable-length numeric, including decimal point if required, suggested
		   maximum length 6 characters
		*/
		DiscountPercent string `xml:"j267"`

		/*
		   Optional and non-repeating.

		   Fixed-length, two numeric digits
		*/
		PriceStatus string `xml:"j266"`

		/*
		   Mandatory in each occurrence of the <Price> composite, and non-repeating.

		   Variable length real number, with explicit decimal point when required,
		   suggested maximum length 12 characters
		*/
		PriceAmount string `xml:"j151"`

		/*
		   Optional and non-repeating.

		   Fixed-length, three letters
		*/
		CurrencyCode string `xml:"j152"`

		/*
		   Optional, and repeatable if a single price applies to multiple countries.

		   Fixed-length, two letters. [Note that ISO 3166-1 specifies that country codes
		   shall be sent as upper case only.]
		*/
		CountryCode string `xml:"b251"`

		/*
		   Optional and non-repeating.

		   One or more variable-length codes, each consisting of upper case letters with
		   or without a hyphen, successive codes being separated by spaces.
		   Suggested maximum length 100 characters.
		*/
		Territory string `xml:"j303"`

		/*
		   Optional and non-repeating.

		   One or more fixed-length codes, each with two upper case letters, successive
		   codes being separated by spaces. Suggested maximum length 300
		   characters. [Note that ISO 3166-1 specifies that country codes shall be sent
		   as upper case only.]
		*/
		CountryExcluded string `xml:"j304"`

		/*
		   Optional and non-repeating.

		   One or more variable-length codes, each consisting of upper case letters with
		   or without a hyphen, successive codes being separated by spaces.
		   Suggested maximum length 100 characters.
		*/
		TerritoryExcluded string `xml:"j308"`

		/*
		   Optional and non-repeating.

		   Fixed-length, one letter.
		*/
		TaxRateCode1 string `xml:"j153"`

		/*
		   A tax rate expressed numerically as a percentage. See notes on <TaxRateCode1>.

		   Variable length real number, with an explicit decimal point where required.
		*/
		TaxRatePercent1 string `xml:"j154"`

		/*
		   This may be the whole of the unit price before tax, if the item carries tax at the same rate on the whole price;
		   or part of the unit price in the case of a mixed tax rate product. See notes on <TaxRateCode1>.

		   Variable length real number, with an explicit decimal point where required.
		*/
		TaxableAmount1 string `xml:"j155"`

		/*
		   The amount of tax chargeable at the rate specified by <TaxRateCode1> and/or
		   <TaxRatePercent1>. See notes on <TaxRateCode1>.

		   Variable length real number, with an explicit decimal point where required.
		*/
		TaxAmount1 string `xml:"j156"`

		/*
		   A code which specifies a value added tax rate applying to the amount of the price which is specified
		   in <TaxableAmount2>. See notes on <TaxRateCode1>.

		   Fixed-length, one letter.
		*/
		TaxRateCode2 string `xml:"j157"`

		/*
		   A tax rate expressed numerically as a percentage. See notes on <TaxRateCode1>.

		   Variable length real number, with an explicit decimal point where required.
		*/
		TaxRatePercent2 string `xml:"j158"`

		/*
		   This may be the whole of the unit price before tax, if
		   the item carries tax at the same rate on the whole price; or part of the unit price in the case of a
		   mixed tax rate product. See notes on <TaxRateCode1>.

		   Variable length real number, with an explicit decimal point where required.
		*/
		TaxableAmount2 string `xml:"j159"`

		/*
		   The amount of tax chargeable at the rate specified by <TaxRateCode2> and/or
		   <TaxRatePercent2>. See notes on <TaxRateCode1>.

		   Variable length real number, with an explicit decimal point where required.
		*/
		TaxAmount2 string `xml:"j160"`

		/*
		   The date from which a price becomes effective. Optional and non-repeating.

		   Date as year, month, day (YYYYMMDD)
		*/
		PriceEffectiveFrom string `xml:"j161"`

		/*
		   The date until which a price remains effective. Optional and non-repeating.

		   Date as year, month, day (YYYYMMDD)
		*/
		PriceEffectiveUntil string `xml:"j162"`
	}

	BatchBonus struct {

		/*
		   Mandatory in each occurrence of the <BatchBonus> composite, and non-repeating.

		   Variable-length integer, suggested maximum length 4 digits
		*/
		BatchQuantity string `xml:"j264"`

		/*
		   Mandatory in each occurrence of the <BatchBonus> composite, and non-repeating.

		   Variable-length integer, suggested maximum length 4 digits
		*/
		FreeQuantity string `xml:"j265"`
	}

	DiscountCoded struct {

		/*
		   Mandatory in each occurrence of the <DiscountCoded> composite, and non-repeating.

		   Fixed-length, 2 numeric digits
		*/
		DiscountCodeType string `xml:"j363"`

		/*
		   Optional and non-repeating.

		   Free text, suggested maximum length 50 characters
		*/
		DiscountCodeTypeName string `xml:"j378"`

		/*
		   Mandatory in each occurrence of the <DiscountCoded> composite, and non-repeating.

		   According to the scheme specified in <DiscountCodeType>
		*/
		DiscountCode string `xml:"j364"`
	}

	Reissue struct {

		/*
		   Mandatory in each occurrence of the <Reissue> composite, and non-repeating.

		   Date as year, month, day (YYYYMMDD)
		*/
		ReissueDate string `xml:"j365"`

		/*
		   Text explaining the nature of the reissue. Optional and non-repeating.

		   Free text, suggested maximum length 500 characters
		*/
		ReissueDescription string `xml:"j366"`

		Price []Price `xml:"price"`

		MediaFile []MediaFile `xml:"mediafile"`
	}

	MarketRepresentation struct {
		AgentIdentifier AgentIdentifier `xml:"agentidentifier"`

		/*
		   Optional and non-repeating; required if no agent identifier is sent in an occurrence
		   of the <MarketRepresentation> composite.

		   Variable-length text, suggested maximum length 100 characters
		*/
		AgentName string `xml:"j401"`

		/*
		   Optional and repeatable.

		   Variable-length text, suggested maximum length 20 characters
		*/
		TelephoneNumber string `xml:"j270"`

		/*
		   A fax number of an agent or local publisher. Optional and repeatable.

		   Variable-length text, suggested maximum length 20 characters
		*/
		FaxNumber string `xml:"j271"`

		/*
		   An email address for an agent or local publisher. Optional and repeatable.

		   Variable-length text, suggested maximum length 100 characters
		*/
		EmailAddress string `xml:"j272"`

		Website []Website `xml:"website"`

		/*
		   Optional and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		AgentRole string `xml:"j402"`

		/*
		   Optional, but each occurrence of the <MarketRepresentation> composite must carry either
		   an occurrence of <MarketCountry> or an occurrence of <MarketTerritory>, to specify the market concerned.
		   Non-repeating.

		   One or more fixed-length codes, each with two upper case letters, successive
		   codes being separated by spaces. Suggested maximum length 600
		   characters. [Note that ISO 3166-1 specifies that country codes shall be sent
		   as upper case only.]
		*/
		MarketCountry string `xml:"j403"`

		/*
		   Optional, but each occurrence of the <MarketRepresentation> composite must carry either an
		   occurrence of <MarketCountry> or an occurrence of <MarketTerritory>, to specify the market
		   concerned. Non-repeating.

		   One or more variable-length codes, each consisting of upper case letters with
		   or without a hyphen, successive codes being separated by spaces.
		   Suggested maximum length 100 characters.
		*/
		MarketTerritory string `xml:"j404"`

		/*
		   Optional and non-repeating.

		   One or more fixed-length codes, each with two upper case letters, successive
		   codes being separated by spaces. Suggested maximum length 300
		   characters. [Note that ISO 3166-1 specifies that country codes shall be sent
		   as upper case only.]
		*/
		MarketCountryExcluded string `xml:"j405"`

		/*
		   Optional and non-repeating.

		   Variable-length text, suggested maximum length 300 characters
		*/
		MarketRestrictionDetail string `xml:"j406"`

		/*
		   Optional and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		MarketPublishingStatus string `xml:"j407"`

		MarketDate MarketDate `xml:"marketdate"`
	}

	AgentIdentifier struct {

		/*
		   Mandatory in each occurrence of the <AgentIdentifier> composite, and non-repeating.

		   Fixed-length, 2 numeric digits
		*/
		AgentIDType string `xml:"j400"`

		/*
		   Optional and non-repeating.

		   Free text, suggested maximum length 50 characters
		*/
		IDTypeName string `xml:"b233"`

		/*
		   Mandatory in each occurrence of the <AgentIdentifier> composite, and non-repeating.

		   According to the identifier type specified in <AgentIDType>
		*/
		IDValue string `xml"b244"`
	}

	MarketDate struct {

		/*
		   Mandatory in each occurrence of the<MarketDate> composite.

		   Fixed-length, two numeric digits
		*/
		MarketDateRole string `xml:"j408"`

		/*
		   Optional and non-repeating. When omitted, the format is assumed to be YYYYMMDD.

		   Fixed-length, two numeric digits
		*/
		DateFormat string `xml:"j260"`

		/*
		   Mandatory in each occurrence of the <MarketDate> composite.

		   As specified by the value in <DateFormat>: default YYYYMMDD
		*/
		Date string `xml:"b306"`
	}

	Header struct {

		/*
		   Optional and non-repeating; but either the <FromCompany> element or a sender identifier using one or more
		   elements from MH.1 to MH.5 must be included.

		   Fixed-length, thirteen numeric digits, of which the last is a check digit.
		*/
		FromEANNumber string `xml:"m172"`

		/*
		   Optional and non-repeating; but either the <FromCompany> element or a sender identifier using
		   one or more elements from MH.1 to MH.5 must be included.

		   Fixed-length, seven characters. The first six are numeric digits, and the
		   seventh is a check character which may be a numeric digit or letter X.
		*/
		FromSAN string `xml:"m173"`

		/*
		   Optional and non-repeating; but either the <FromCompany> element or a sender
		   identifier using one or more elements from MH.1 to MH.5 must be included.

		   Variable-length ASCII text, suggested maximum 30 characters
		*/
		FromCompany string `xml:"m174"`

		/*
		   Optional and non-repeating.

		   Variable-length ASCII text, suggested maximum 300 characters
		*/
		FromPerson string `xml:"m175"`

		/*
		   Optional and non-repeating.

		   Variable-length ASCII text, suggested maximum 100 characters
		*/
		FromEmail string `xml:"m283"`

		/*
		   Optional and non-repeating.

		   Fixed-length, thirteen numeric digits, of which the last is a check digit.
		*/
		ToEANNumber string `xml:"m176"`

		/*
		   Optional and non-repeating.

		   Fixed-length, seven characters. The first six are numeric digits, and the
		   seventh is a check character which may be a numeric digit or letter X.
		*/
		ToSAN string `xml:"m177"`

		/*
		   Optional and non-repeating.

		   Variable-length ASCII text, suggested maximum 30 characters
		*/
		ToCompany string `xml:"m178"`

		/*
		   Optional and non-repeating.

		   Variable-length ASCII text, suggested maximum 300 characters
		*/
		ToPerson string `xml:"m179"`

		/*
		   Optional and non-repeating.

		   Variable-length integer,
		*/
		MessageNumber string `xml:"m180"`

		/*
		   Optional and non-repeating.

		   Variable-length intege
		*/
		MessageRepeat string `xml:"m181"`

		/*
		   Mandatory and non-repeating.

		   Eight or twelve numeric digits only (YYYYMMDD or YYYYMMDDHHMM)
		*/
		SentDate string `xml:"m182"`

		/*
		   Optional and non-repeating.

		   Variable-length ASCII text, suggested maximum 500 characters
		*/
		MessageNote string `xml:"m183"`

		/*
		   Optional and non-repeating.

		   Fixed-length, three letters.
		*/
		DefaultLanguageOfText string `xml:"m184"`

		/*
		   Optional and non-repeating.

		   Fixed-length, two numeric digits.
		*/
		DefaultPriceTypeCode string `xml:"m185"`

		/*
		   Optional and non-repeating.

		   Fixed-length, three letters.
		*/
		DefaultCurrencyCode string `xml:"m186"`

		/*
		   Optional and non-repeating.

		   ASCII text, suggested maximum length 50 characters.
		*/
		DefaultClassOfTrade string `xml:"m193"`
	}

	SenderIdentifier struct {

		/*      Mandatory in any occurrence of the <SenderIdentifier> composite, and non-repeating.

		        Fixed-length, two numeric digits
		*/
		SenderIDType string `xml:"m379"`

		/*
		   Optional and non-repeating.

		   Variable-length ASCII text, suggested maximum 50 characters
		*/
		IDTypeName string `xml:"b233"`

		/*
		   Mandatory in any occurrence of the <SenderIdentifier> composite, and non-repeating.

		   According to the identifier type specified in <SenderIDType>
		*/
		IDValue string `xml:"b244"`
	}

	AddresseeIdentifier struct {

		/*
		   Mandatory in any occurrence of the <AddresseeIdentifier> composite, and non-repeating.

		   Fixed-length, two numeric digits
		*/
		AddresseeIDType string `xml:"m380"`

		/*
		   Optional and non-repeating.

		   Variable-length ASCII text, suggested maximum 50 characters
		*/
		IDTypeName string `xml:"b233"`

		/*
		   Mandatory in any occurrence of the <AddresseeIdentifier> composite, and non-repeating.

		   According to the identifier type specified in <AddresseeIDType>
		*/
		IDValue string `xml:"b244"`
	}
)

const (
	rootPath   = "/Onix"
	rootPathDL = "/OnixDL"

	bzFileName          = "BzTransferFull"
	bzDLFileName        = "BzTransferFullDL"
	bzPartialFileName   = "bzonix"
	bzPartialDLFileName = "bzonixdl"
)

func loadOnixFiles(syncData model.Sync, zipLocation, saveLocation string, isDL bool) (newestOnix int64, err error) {
	log := logger.Log.WithOptions(zap.Fields(
		zap.String("zipLocation", zipLocation),
		zap.Any("saveLocation", saveLocation),
	))
	log.Info("started load onix files")

	rootLoc := rootPath
	bzFileLoc := bzFileName
	bzPartialLoc := bzPartialFileName
	if isDL {
		rootLoc = rootPathDL
		bzFileLoc = bzDLFileName
		bzPartialLoc = bzPartialDLFileName
	}

	if !syncData.IsFullSynced {
		syncData.LastOnixSyncDate, err = FindFullOnixFiles(rootLoc, bzFileLoc, zipLocation, saveLocation)
		if err != nil {
			log.Error("failed to find full onix files", zap.Error(err))
			return 0, err
		}
	}
	log.Info("started load partial onix files")

	syncData.LastOnixSyncDate, err = FindPartialOnixFiles(rootLoc, bzPartialLoc, zipLocation, saveLocation, syncData.LastOnixSyncDate)
	if err != nil {
		log.Error("failed to find partial onix files", zap.Error(err))
		return 0, err
	}

	return syncData.LastOnixSyncDate, nil
}

func FindFullOnixFiles(ftpRootPath, bzFile, zipLocation, saveLocation string) (newestDate int64, err error) {
	// FTP server URL
	ftpClient, err := ConnectToFtp(ftpUrl)
	if err != nil {
		logger.Log.Error("failed to connect to ftp",
			zap.Error(err),
		)
		return
	}
	defer ftpClient.Close()

	entries, err := ftpClient.ReadDir(ftpRootPath)
	if err != nil {
		logger.Log.Error("failed to list ftp server", zap.Error(err))
		return
	}

	var (
		index = -1
	)

	logger.Log.Info("loading entries")
	for i, entry := range entries {
		if strings.Contains(entry.Name(), "BZtransferFullArchive") {
			continue
		}

		if !strings.Contains(entry.Name(), bzFile) {
			continue
		}

		splits := strings.Split(entry.Name(), bzFile)
		if len(splits) < 2 {
			err = fmt.Errorf("failed to split entry name")
			return
		}

		splits = strings.Split(splits[1], ".")
		if len(splits) == 0 {
			err = fmt.Errorf("incorrect onix file name")
			return
		}

		layout := "20060102"
		t, err := time.Parse(layout, splits[0])
		if err != nil {
			return 0, err
		}

		if t.Unix() > newestDate {
			newestDate = t.Unix()
			index = i
		}
	}

	if len(entries) == 0 {
		err = fmt.Errorf("failed to list entries")
		return
	}

	if index == -1 {
		err = fmt.Errorf("unable to find valid file")
		return
	}

	logger.Log.Info("found newest entry",
		zap.String("entryName", entries[index].Name()),
	)

	filePath := filepath.Join(ftpRootPath, entries[index].Name())
	err = DownloadAndDecompress(ftpClient, filePath, zipLocation, saveLocation)
	if err != nil {
		return
	}

	return
}

func FindPartialOnixFiles(ftpRootPath, bzFile, zipLocation, saveLocation string, lastSyncDate int64) (newestDate int64, err error) {
	// FTP server URL
	ftpClient, err := ConnectToFtp(ftpUrl)
	if err != nil {
		logger.Log.Error("failed to connect to ftp",
			zap.Error(err),
		)
		return
	}
	defer ftpClient.Close()

	entries, err := ftpClient.ReadDir(rootPath)
	if err != nil {
		logger.Log.Error("failed to list ftp server", zap.Error(err))
		return
	}

	var (
		index int
	)

	logger.Log.Info("loading entries")
	for i, entry := range entries {
		if !strings.Contains(entry.Name(), bzFile) {
			continue
		}

		splits := strings.Split(entry.Name(), bzFile)
		if len(splits) < 2 {
			err = fmt.Errorf("failed to split entry name")
			return
		}

		splits = strings.Split(splits[1], ".")
		if len(splits) == 0 {
			err = fmt.Errorf("incorrect onix file name")
			return
		}

		layout := "20060102"
		t, err := time.Parse(layout, splits[0])
		if err != nil {
			return 0, err
		}

		if t.Unix() > newestDate {
			newestDate = t.Unix()
			index = i
		}

		if t.Unix() > lastSyncDate {
			filePath := filepath.Join(ftpRootPath, entry.Name())

			err = DownloadAndDecompress(ftpClient, filePath, zipLocation, saveLocation)
			if err != nil {
				return 0, err
			}
		}
	}

	if len(entries) == 0 {
		err = fmt.Errorf("failed to list entries")
		return
	}

	logger.Log.Info("found newest entry",
		zap.String("entryName", entries[index].Name()),
	)

	return
}

func loadOnixData(dataLocation string, isDL bool) (err error) {
	defer os.RemoveAll(dataLocation)

	logger.Log.Info("reading xml files from disk",
		zap.String("directoryPath", dataLocation),
	)

	files, err := os.ReadDir(dataLocation)
	if err != nil {
		logger.Log.Info("already up to date",
			zap.Error(err),
		)
		err = nil
		return
	}

	for _, file := range files {
		onix := Onix{}
		filePath := filepath.Join(dataLocation, file.Name())

		log := logger.Log.WithOptions(zap.Fields(
			zap.Any("filePath", filePath),
		))

		raw, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		log.Info("getting data from xml for products")

		err = xml.Unmarshal(raw, &onix)
		if err != nil {
			return err
		}

		productCreate := make([]model.Product, 0)
		productUpdate := make([]model.Product, 0)
		for _, product := range onix.Products {
			p, err := parseProductData(product, isDL)
			if err != nil {
				log.Debug("skipping product",
					zap.Error(err),
				)
				continue
			}

			// TODO:
			if isDL {
				p.IsDownloadTitle = true
			}

			var product model.Product
			if err := database.DB.Where("id = ?", p.ID).First(&product).Error; err != nil {
				p.SalesChannels = append(p.SalesChannels, model.SalesChannelProduct{
					SalesChannelID: "1",
				})

				productCreate = append(productCreate, p)
			} else {
				productUpdate = append(productUpdate, p)
			}
		}

		log.Info("uploading batch products",
			zap.Int("updateCount", len(productUpdate)),
			zap.Int("createCount", len(productCreate)),
		)

		err = batchCreate(productCreate, 100)
		if err != nil {
			log.Error("failed batch create",
				zap.Error(err),
			)
			return err
		}

		wg := sync.WaitGroup{}
		wg.Add(2)
		err = batchUpdate(productUpdate[0:len(productUpdate)/2], &wg)
		if err != nil {
			log.Error("failed batch update",
				zap.Error(err),
			)

			return err
		}

		err = batchUpdate(productUpdate[len(productUpdate)/2:], &wg)
		if err != nil {
			log.Error("failed batch update",
				zap.Error(err),
			)

			return err
		}
		wg.Wait()
	}

	return
}

func formatISBN(isbn string) string {
	// Define the regular expression pattern for matching the ISBN
	pattern := regexp.MustCompile(`^(\d{3})(\d{1})(\d{5})(\d{3})(\d{1})$`)

	// Apply the pattern and capture the groups
	matches := pattern.FindStringSubmatch(isbn)
	if len(matches) != 6 {
		return isbn // Return the original ISBN if the pattern does not match
	}

	// Build the formatted ISBN string
	formattedISBN := fmt.Sprintf("%s-%s-%s-%s-%s", matches[1], matches[2], matches[3], matches[4], matches[5])

	return formattedISBN
}

func parseProductData(input Product, isDL bool) (output model.Product, err error) {
	// author
	hasAuthor := false
	maximum := 5
	for _, c := range input.Contributor {
		if c.ContributorRole == "A01" {
			if !hasAuthor {
				output.Publisher = c.PersonNameInverted
			} else {
				output.Publisher += ";" + c.PersonNameInverted
			}

			maximum--
			hasAuthor = true
		}

		if maximum == 0 {
			break
		}
	}

	// language
	language := ""
	for _, lang := range input.Language {
		language = language + languageMap[lang.LanguageCode] + " / "
	}

	language = strings.Trim(language, " / ")
	output.Language = language

	// edition
	output.Edition = input.EditionStatement

	// stock if unavailable then stock 0
	output.Stock = 1000
	if input.PublishingStatus == "08" {
		output.Stock = 0
	}

	// delivery time. If delivery time is missing stock needs to be 0
	splits := strings.Split(input.PublishingStatusNote, ":")
	if len(splits) == 2 {
		output.DeliveryTime = splits[1]
		if !strings.Contains(output.DeliveryTime, "Lieferbar in") {
			output.Stock = 0
		}
	} else if !isDL {
		output.Stock = 0
	}

	// publication date
	layout := "20060102"
	if len(input.PublicationDate) == 4 {
		layout = "2006"
	} else if len(input.PublicationDate) == 6 {
		layout = "200601"
	}

	t, err := time.Parse(layout, input.PublicationDate)
	if err == nil {
		output.PublicationDate = t.Unix()
	}

	// measure
	for _, m := range input.Measure {
		if m.MeasureTypeCode == "01" {
			output.Height = m.Measurement
		}

		if m.MeasureTypeCode == "02" {
			output.Width = m.Measurement
		}

		if m.MeasureTypeCode == "03" {
			output.Length = m.Measurement
		}

		if m.MeasureTypeCode == "08" {
			output.Weight = m.Measurement
		}
	}

	// izbn, ean, bz number
	for _, productIdentifier := range input.ProductIdentifier {
		// PRODUCT NUMMER AND ISBN
		switch productIdentifier.ProductIDType {
		case "01":
			switch productIdentifier.IDTypeName {
			case "01":
				output.ID = productIdentifier.IDValue
				output.BZNR = productIdentifier.IDValue
			case "03":
				// PRODUCT NUMBER
			}
		case "03":
			output.EAN = productIdentifier.IDValue
		case "15":
			output.ISBN = formatISBN(productIdentifier.IDValue)
		}
	}

	if strings.TrimSpace(output.ISBN) == "" {
		output.ISBN = output.EAN
	}

	if output.ID == "" {
		output.ID = input.RecordReference
		if output.ID == "" {
			err = fmt.Errorf("skipping due to bz number not existing")
			return
		}
	}

	// price
	for _, supply := range input.SupplyDetail {
		for _, p := range supply.Price {
			if p.PriceTypeCode == "02" {
				output.SellingPrice, err = strconv.ParseFloat(p.PriceAmount, 64)
				if err != nil {
					return
				}
			}
		}
	}

	if output.SellingPrice == 0 {
		err = fmt.Errorf("skipping due to selling price being 0")
		return
	}

	// title and subtitle
	for _, t := range input.Title {
		output.Title = t.TitleText
		output.Subtitle = t.Subtitle
		if t.TitleType == "01" {
			break
		}
	}

	// category
	for _, subject := range input.MainSubject {
		if subject.MainSubjectSchemeIdentifier != "26" {
			continue
		}

		const (
			DVDVideoCategory       = "1001"
			AudioCDCategory        = "1002"
			CDROMCategory          = "1003"
			CalendarCategory       = "1004"
			MapsCategory           = "1005"
			NonBooksCategory       = "1006"
			EbooksCategory         = "1007"
			AudioDownloadsCategory = "1020"
		)
		categoryId := subject.SubjectCode[1:]

		if val, found := newCategoriesMap[categoryId]; found {
			var category model.Category
			res := database.DB.First(&category, val)
			if res.RowsAffected != 0 {
				output.Categories = make([]model.ProductCategory, 0)
				output.Categories = append(output.Categories, model.ProductCategory{
					CategoryID: subject.SubjectCode,
				})
			}
		}

		if isDL {
			switch input.ProductForm {
			case "DG":
				var category model.Category
				res := database.DB.First(&category, EbooksCategory)
				if res.RowsAffected == 0 {
					err = fmt.Errorf("skipping due to category not existing")
					return
				}

				output.Categories = make([]model.ProductCategory, 0)
				output.Categories = append(output.Categories, model.ProductCategory{
					CategoryID: EbooksCategory,
				})

				if val, found := categoryMap[categoryId]; found {
					var category model.Category
					res := database.DB.First(&category, val.EbookId)
					if res.RowsAffected == 0 {
						err = fmt.Errorf("skipping due to category not existing")
						return
					}

					output.Categories = append(output.Categories, model.ProductCategory{
						CategoryID: val.EbookId,
					})
				}
			case "AJ":
				var category model.Category
				res := database.DB.First(&category, AudioDownloadsCategory)
				if res.RowsAffected == 0 {
					err = fmt.Errorf("skipping due to category not existing")
					return
				}

				output.Categories = make([]model.ProductCategory, 0)
				output.Categories = append(output.Categories, model.ProductCategory{
					CategoryID: AudioDownloadsCategory,
				})

				if val, found := categoryMap[categoryId]; found {
					var category model.Category
					res := database.DB.First(&category, val.AudioDownloadId)
					if res.RowsAffected == 0 {
						err = fmt.Errorf("skipping due to category not existing")
						return
					}

					output.Categories = append(output.Categories, model.ProductCategory{
						CategoryID: val.AudioDownloadId,
					})
				}
			default:
				err = fmt.Errorf("download title is not DG/AJ")
				return
			}

		} else {
			var firstChar byte = subject.SubjectCode[0]
			switch firstChar {
			case '1':
				subject.SubjectCode = subject.SubjectCode[1:]
			case '2':
				subject.SubjectCode = subject.SubjectCode[1:]
			case '3':
				subject.SubjectCode = subject.SubjectCode[1:]
			case '4':
				subject.SubjectCode = DVDVideoCategory
			case '5':
				subject.SubjectCode = AudioCDCategory
			case '6':
				subject.SubjectCode = CDROMCategory
			case '7':
				subject.SubjectCode = CalendarCategory
			case '8':
				subject.SubjectCode = MapsCategory
			case '9':
				subject.SubjectCode = NonBooksCategory
			default:
				err = fmt.Errorf("encountered invalid first char %c", firstChar)
				return
			}

			var category model.Category
			res := database.DB.First(&category, subject.SubjectCode)
			if res.RowsAffected == 0 {
				err = fmt.Errorf("skipping due to category not existing")
				return
			}

			output.Categories = make([]model.ProductCategory, 0)
			output.Categories = append(output.Categories, model.ProductCategory{
				CategoryID: subject.SubjectCode,
			})

			// add to andere and bucher root category
			switch firstChar {
			case '9':
				output.Categories = append(output.Categories, model.ProductCategory{
					CategoryID: "9999",
				})

			case '1':
				output.Categories = append(output.Categories, model.ProductCategory{
					CategoryID: "90",
				})
			}
		}
	}

	if len(input.RelatedProduct) != 0 {
		for _, productIdentifier := range input.RelatedProduct[0].ProductIdentifier {
			if productIdentifier.ProductIDType == "01" && productIdentifier.IDTypeName == "01" {
				output.Replacement = productIdentifier.IDValue
			}
		}
	}

	return
}

type CategoryMap struct {
	EbookId         string
	AudioDownloadId string
}

var (
	languageMap = make(map[string]string)
	//go:embed langMap.txt
	langMapRaw []byte

	categoryMap = make(map[string]CategoryMap)
	//go:embed downloadTitleMap.txt
	categoryMapRaw []byte

	newCategoriesMap = make(map[string]string)
	//go:embed newCategoriesMap.txt
	newcategoriesMapRaw []byte
)

func init() {
	splits := strings.Split(string(langMapRaw), "\n")
	for _, split := range splits {
		s := strings.Split(split, " ")
		if len(splits) < 2 {
			continue
		}

		languageMap[s[0]] = strings.Trim(strings.Join(s[1:], " "), " ")
	}

	splits = strings.Split(string(categoryMapRaw), "\n")
	for _, split := range splits {
		s := strings.Split(split, " ")
		s2 := strings.Split(s[1], ",")
		categoryMap[s[0]] = CategoryMap{
			EbookId:         strings.Trim(s2[0], " "),
			AudioDownloadId: strings.Trim(s2[1], " "),
		}
	}

	splits = strings.Split(string(newcategoriesMapRaw), "\n")
	for _, split := range splits {
		s := strings.Split(split, " ")
		if len(splits) < 2 {
			continue
		}

		newCategoriesMap[s[0]] = strings.Trim(strings.Join(s[1:], " "), " ")
	}
}
