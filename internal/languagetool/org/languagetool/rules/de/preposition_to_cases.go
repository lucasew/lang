package de

import "strings"

// GrammaticalCase is PrepositionToCases.Case (Nom, Akk, Gen, Dat).
// Twin of org.languagetool.rules.de.PrepositionToCases.Case.
type GrammaticalCase int

const (
	CaseNom GrammaticalCase = iota // Nom
	CaseAkk                        // Akk
	CaseGen                        // Gen
	CaseDat                        // Dat
)

// PrepositionCases is the prep2Cases map from PrepositionToCases static init.
// Keys and case order match Java add(...) calls; commented-out Java adds are absent.
// Twin of org.languagetool.rules.de.PrepositionToCases.prep2Cases.
var PrepositionCases = map[string][]GrammaticalCase{
	"ab":             {CaseDat},
	"abseits":        {CaseGen},
	"abzüglich":      {CaseGen},
	"an":             {CaseDat, CaseAkk},
	"angesichts":     {CaseGen},
	"anhand":         {CaseGen},
	"anlässlich":     {CaseGen},
	"anstatt":        {CaseGen},
	"anstelle":       {CaseGen},
	"auf":            {CaseDat, CaseAkk},
	"aufgrund":       {CaseGen},
	"aus":            {CaseDat},
	"ausgangs":       {CaseGen},
	"ausgenommen":    {CaseAkk},
	"ausschließlich": {CaseGen},
	"ausweislich":    {CaseGen},
	"außer":          {CaseDat},
	"außerhalb":      {CaseGen},
	"bar":            {CaseGen},
	"behufs":         {CaseGen},
	"bei":            {CaseDat},
	"beidseits":      {CaseGen},
	"betreffend":     {CaseAkk},
	"betreffs":       {CaseGen},
	"bezüglich":      {CaseGen},
	"binnen":         {CaseDat},
	"bis":            {CaseAkk},
	"bis auf":        {CaseAkk},
	"dank":           {CaseDat, CaseGen},
	"diesseits":      {CaseGen},
	"durch":          {CaseAkk},
	"eingangs":       {CaseGen},
	"eingedenk":      {CaseGen},
	"einschließlich": {CaseGen},
	"entgegen":       {CaseDat},
	"entlang":        {CaseDat, CaseAkk, CaseGen},
	"entsprechend":   {CaseDat},
	"exklusive":      {CaseGen},
	"fern":           {CaseDat},
	"fernab":         {CaseGen},
	"für":            {CaseAkk},
	"gegen":          {CaseAkk}, // Java: add("gegen", /*Dat,*/ Akk)
	"gegenüber":      {CaseDat},
	"gemäß":          {CaseDat},
	"getreu":         {CaseDat},
	"halber":         {CaseGen},
	"hinsichtlich":   {CaseGen},
	"hinter":         {CaseDat, CaseAkk},
	// Java: //add("in", Dat, Akk); deliberately absent
	"in puncto":      {CaseGen},
	"infolge":        {CaseGen},
	"inklusive":      {CaseGen},
	"inmitten":       {CaseGen},
	"innerhalb":      {CaseGen},
	"innert":         {CaseGen},
	"je":             {CaseAkk},
	"jenseits":       {CaseGen},
	"kontra":         {CaseAkk},
	"kraft":          {CaseGen},
	"laut":           {CaseGen, CaseDat},
	"längs":          {CaseGen},
	"längsseits":     {CaseGen},
	"mangels":        {CaseGen},
	"mit":            {CaseDat},
	"mitsamt":        {CaseDat},
	"mittels":        {CaseGen},
	"nach":           {CaseDat},
	"namens":         {CaseGen},
	"neben":          {CaseDat, CaseAkk},
	"nebst":          {CaseDat},
	"nächst":         {CaseDat},
	// Java: //add("ob", Gen); deliberately absent
	"oberhalb":       {CaseGen},
	"ohne":           {CaseAkk},
	"per":            {CaseAkk},
	"pro":            {CaseAkk},
	"rücksichtlich":  {CaseGen},
	"samt":           {CaseDat},
	"seit":           {CaseDat},
	"seitens":        {CaseGen},
	"seitlich":       {CaseGen},
	"statt":          {CaseGen},
	"trotz":          {CaseGen},
	"um":             {CaseAkk},
	"unbeschadet":    {CaseGen},
	"uneingedenk":    {CaseGen},
	"ungeachtet":     {CaseGen},
	"unter":          {CaseDat, CaseAkk},
	"unterhalb":      {CaseGen},
	"unweit":         {CaseGen},
	"vermittels":     {CaseGen},
	"vermittelst":    {CaseGen},
	"vermöge":        {CaseGen},
	"versus":         {CaseAkk},
	"via":            {CaseAkk},
	"vis-à-vis":      {CaseDat},
	"von":            {CaseDat},
	"vor":            {CaseDat, CaseAkk},
	"vorbehaltlich":  {CaseGen},
	"wegen":          {CaseGen},
	"weitab":         {CaseGen},
	"wider":          {CaseAkk},
	"während":        {CaseGen},
	"zeit":           {CaseGen},
	"zu":             {CaseDat},
	"zufolge":        {CaseDat},
	"zugunsten":      {CaseGen},
	"zuliebe":        {CaseDat},
	"zuungunsten":    {CaseGen},
	"zuwider":        {CaseDat},
	"zuzüglich":      {CaseGen},
	"zwecks":         {CaseGen},
	"zwischen":       {CaseDat, CaseAkk},
	"à":              {CaseAkk},
	"über":           {CaseDat, CaseAkk},
}

// CasesForPreposition ports PrepositionToCases.getCasesFor.
// Returns the case list for prep (lowercased), or an empty non-nil slice when unknown
// (Java: prep2Cases.getOrDefault(prep.toLowerCase(), new ArrayList<>())).
func CasesForPreposition(prep string) []GrammaticalCase {
	if c, ok := PrepositionCases[strings.ToLower(prep)]; ok {
		return c
	}
	return []GrammaticalCase{}
}
