// WIP.
package exporter

import (
	"encoding/json"

	"github.com/kennygrant/sanitize"
	"github.com/miku/span/container"
	"github.com/miku/span/finc"
)

// Attach attaches the ISILs to a record. Noop.
func (s *Solr5Vufind3v12) Attach(_ []string) {}

// WIP: Solr5Vufind3v12 is the basic solr 5 schema as of 2016-04-14. It is based on
// VuFind 3. Same as Solr5Vufind3v12, but with fullrecord field, refs. #8031.

type Solr5Vufind3v12 struct {
	AccessFacet          string   `json:"access_facet,omitempty"`
	AuthorFacet          []string `json:"author_facet,omitempty"`
	Authors              []string `json:"author,omitempty"`
	SecondaryAuthors     []string `json:"author2,omitempty"`
	Allfields            string   `json:"allfields,omitempty"`
	FincClassFacet       []string `json:"finc_class_facet,omitempty"`
	Formats              []string `json:"format,omitempty"`
	Fullrecord           string   `json:"fullrecord,omitempty"`
	Fulltext             string   `json:"fulltext,omitempty"`
	HierarchyParentTitle []string `json:"hierarchy_parent_title,omitempty"`
	ID                   string   `json:"id,omitempty"`
	Institutions         []string `json:"institution,omitempty"`
	Imprint              string   `json:"imprint,omitempty"`
	ISSN                 []string `json:"issn,omitempty"`
	Languages            []string `json:"language,omitempty"`
	MegaCollections      []string `json:"mega_collection,omitempty"`
	PublishDateSort      int      `json:"publishDateSort,omitempty"`
	Publishers           []string `json:"publisher,omitempty"`
	RecordType           string   `json:"recordtype,omitempty"`
	Series               []string `json:"series,omitempty"`
	SourceID             string   `json:"source_id,omitempty"`
	Subtitle             string   `json:"title_sub,omitempty"`
	Title                string   `json:"title,omitempty"`
	TitleFull            string   `json:"title_full,omitempty"`
	TitleShort           string   `json:"title_short,omitempty"`
	TitleSort            string   `json:"title_sort,omitempty"`
	Topics               []string `json:"topic,omitempty"`
	URL                  []string `json:"url,omitempty"`
	PublishDate          []string `json:"publishDate,omitempty"`

	VF1Author           string   `json:"vf1_author,omitempty"`
	VF1SecondaryAuthors []string `json:"vf1_author2,omitempty"`

	ContainerIssue     string `json:"container_issue,omitempty"`
	ContainerStartPage string `json:"container_start_page,omitempty"`
	ContainerTitle     string `json:"container_title,omitempty"`
	ContainerVolume    string `json:"container_volume,omitempty"`

	FormatDe105  []string `json:"format_de105,omitempty"`
	FormatDe14   []string `json:"format_de14,omitempty"`
	FormatDe15   []string `json:"format_de15,omitempty"`
	FormatDe520  []string `json:"format_de520,omitempty"`
	FormatDe540  []string `json:"format_de540,omitempty"`
	FormatDeCh1  []string `json:"format_dech1,omitempty"`
	FormatDed117 []string `json:"format_ded117,omitempty"`
	FormatDeGla1 []string `json:"format_degla1,omitempty"`
	FormatDel152 []string `json:"format_del152,omitempty"`
	FormatDel189 []string `json:"format_del189,omitempty"`
	FormatDeZi4  []string `json:"format_dezi4,omitempty"`
	FormatDeZwi2 []string `json:"format_dezwi2,omitempty"`
}

// Export method from intermediate schema to solr 4/13 schema.
func (s *Solr5Vufind3v12) Convert(is finc.IntermediateSchema) error {
	s.Allfields = is.Allfields()
	s.Formats = append(s.Formats, is.Format)
	s.Fullrecord = "blob:" + is.RecordID
	s.Fulltext = is.Fulltext
	s.ID = is.RecordID
	s.Imprint = is.Imprint()
	s.ISSN = is.ISSNList()
	s.MegaCollections = append(s.MegaCollections, is.MegaCollection)
	s.PublishDateSort = is.Date.Year()
	s.PublishDate = []string{is.Date.Format("2006-01-02")}
	s.Publishers = is.Publishers
	s.RecordType = finc.AIRecordType
	s.Series = append(s.Series, is.JournalTitle)
	s.SourceID = is.SourceID
	s.Subtitle = is.ArticleSubtitle
	s.TitleSort = is.SortableTitle()
	s.Topics = is.Subjects
	s.URL = is.URL

	classes := container.NewStringSet()
	for _, s := range is.Subjects {
		for _, class := range SubjectMapping.LookupDefault(s, []string{}) {
			classes.Add(class)
		}
	}
	s.FincClassFacet = classes.Values()

	sanitized := sanitize.HTML(is.ArticleTitle)
	s.Title, s.TitleFull, s.TitleShort = sanitized, sanitized, sanitized

	for _, lang := range is.Languages {
		s.Languages = append(s.Languages, LanguageMap.LookupDefault(lang, lang))
	}

	// collect sanizized authors
	var authors []string
	for _, author := range is.Authors {
		sanitized := AuthorReplacer.Replace(author.String())
		if sanitized == "" {
			continue
		}
		authors = append(authors, sanitized)

		// first, random author goes into author field, others into secondary field, refs. #5778
		if s.VF1Author == "" {
			s.VF1Author = sanitized
		} else {
			s.VF1SecondaryAuthors = append(s.VF1SecondaryAuthors, sanitized)
		}
		s.AuthorFacet = append(s.AuthorFacet, sanitized)
	}

	if s.VF1Author == "" {
		s.VF1Author = finc.NOT_ASSIGNED
	}

	if len(authors) == 0 {
		s.Authors = []string{finc.NOT_ASSIGNED}
	} else {
		s.Authors = authors
	}

	s.AccessFacet = AIAccessFacet

	// site specific formats
	s.FormatDe105 = []string{FormatDe105.LookupDefault(is.Format, "")}
	s.FormatDe14 = []string{FormatDe14.LookupDefault(is.Format, "")}
	s.FormatDe15 = []string{FormatDe15.LookupDefault(is.Format, "")}
	s.FormatDe520 = []string{FormatDe520.LookupDefault(is.Format, "")}
	s.FormatDe540 = []string{FormatDe540.LookupDefault(is.Format, "")}
	s.FormatDeCh1 = []string{FormatDeCh1.LookupDefault(is.Format, "")}
	s.FormatDed117 = []string{FormatDed117.LookupDefault(is.Format, "")}
	s.FormatDeGla1 = []string{FormatDeGla1.LookupDefault(is.Format, "")}
	s.FormatDel152 = []string{FormatDel152.LookupDefault(is.Format, "")}
	s.FormatDel189 = []string{FormatDel189.LookupDefault(is.Format, "")}
	s.FormatDeZi4 = []string{FormatDeZi4.LookupDefault(is.Format, "")}
	s.FormatDeZwi2 = []string{FormatDeZwi2.LookupDefault(is.Format, "")}

	s.ContainerVolume = is.Volume
	s.ContainerIssue = is.Issue
	s.ContainerStartPage = is.StartPage
	s.ContainerTitle = is.JournalTitle

	s.Institutions = is.Labels

	// refs. #8031
	b, err := json.Marshal(is)
	if err != nil {
		return err
	}
	s.Fullrecord = string(b)
	return nil
}
