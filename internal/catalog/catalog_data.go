package catalog

// CanonicalAcademicUnitSeed is the user-supplied initial NITC catalog.
type CanonicalAcademicUnitSeed struct {
	Code string
	Name string
	Type string
}

var CanonicalAcademicUnitSeeds = []CanonicalAcademicUnitSeed{
	{Code: "BARCH", Name: "B. Arch", Type: "academic_programme"},
	{Code: "BT", Name: "Biotechnology", Type: "department"},
	{Code: "CHE", Name: "Chemical Engineering", Type: "department"},
	{Code: "CE", Name: "Civil Engineering", Type: "department"},
	{Code: "CSE", Name: "Computer Science and Engineering", Type: "department"},
	{Code: "EEE", Name: "Electrical and Electronics Engineering", Type: "department"},
	{Code: "ECE", Name: "Electronics and Communication Engineering", Type: "department"},
	{Code: "ENE", Name: "Energy Engineering", Type: "department"},
	{Code: "EP", Name: "Engineering Physics", Type: "department"},
	{Code: "HSS", Name: "Humanities and Social Sciences", Type: "department"},
	{Code: "MSE", Name: "Materials Science and Engineering", Type: "department"},
	{Code: "ME", Name: "Mechanical Engineering", Type: "department"},
	{Code: "PE", Name: "Production Engineering", Type: "department"},
	{Code: "ITEP", Name: "4-year Integrated Teacher Education Programme (ITEP) B.Sc–B.Ed", Type: "academic_programme"},
}
