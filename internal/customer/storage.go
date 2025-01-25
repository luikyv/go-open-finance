package customer

type Storage struct {
	personalIdentificationsMap    map[string][]PersonalIdentification
	personalQualificationsMap     map[string]PersonalQualifications
	personalFinancialRelationsMap map[string]PersonalFinancialRelations
}

func NewStorage() *Storage {
	return &Storage{
		personalIdentificationsMap:    make(map[string][]PersonalIdentification),
		personalQualificationsMap:     make(map[string]PersonalQualifications),
		personalFinancialRelationsMap: make(map[string]PersonalFinancialRelations),
	}
}

func (s *Storage) addPersonalIdentification(sub string, identification PersonalIdentification) {
	s.personalIdentificationsMap[sub] = append(s.personalIdentificationsMap[sub], identification)
}

func (s *Storage) personalIdentifications(sub string) []PersonalIdentification {
	return s.personalIdentificationsMap[sub]
}

func (s *Storage) setPersonalQualification(sub string, q PersonalQualifications) {
	s.personalQualificationsMap[sub] = q
}

func (s *Storage) personalQualifications(sub string) PersonalQualifications {
	return s.personalQualificationsMap[sub]
}

func (s *Storage) setPersonalFinancialRelations(sub string, rels PersonalFinancialRelations) {
	s.personalFinancialRelationsMap[sub] = rels
}

func (s *Storage) personalFinancialRelations(sub string) PersonalFinancialRelations {
	return s.personalFinancialRelationsMap[sub]
}
