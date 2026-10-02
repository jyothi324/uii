package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func (sd *ServiceDetails) UnmarshalJSON(data []byte) error {
    // Use an alias to avoid recursion
    type Alias ServiceDetails
    aux := &struct {
        ServiceCategory json.RawMessage `json:"service_category"`
        *Alias
    }{
        Alias: (*Alias)(sd),
    }

    if err := json.Unmarshal(data, aux); err != nil {
        return err
    }

    if aux.ServiceCategory == nil || string(aux.ServiceCategory) == "null" {
        sd.ServiceCategory = []ServiceCategory{}
        return nil
    }

    var cats []ServiceCategory
    if err := json.Unmarshal(aux.ServiceCategory, &cats); err == nil {
        sd.ServiceCategory = cats
        return nil
    }

    // Fall back to single object
    var single ServiceCategory
    if err := json.Unmarshal(aux.ServiceCategory, &single); err == nil {
        if single.ID != 0 {
            sd.ServiceCategory = []ServiceCategory{single}
        } else {
            sd.ServiceCategory = []ServiceCategory{}
        }
        return nil
    }

    sd.ServiceCategory = []ServiceCategory{}
    return nil
}

type Service struct {
	ID      uint           `gorm:"primaryKey;"`
	Details ServiceDetails `json:"details" gorm:"type:jsonb"`
}

type ServiceDetails struct {
	Name                string          `gorm:"type:text;"`
	BasicUOM            string          `gorm:"type:text;"`
	SKU                 string          `gorm:"type:text;"`
	ServiceCategory     []ServiceCategory `json:"service_category" gorm:"type:jsonb"`
	// ServiceType         ServiceType     `json:"service_type" gorm:"type:jsonb"`//not using
	ServiceGroup        ServiceGroup    `json:"service_Group" gorm:"type:jsonb"`
	SalesAccount        string          `gorm:"type:text;"`
	AvgUnitCost         uint//not using
	ProAcc              float32
	RetailAcc           float32
	Notes               string `gorm:"type:text;"`
	SpecialInstructions string `gorm:"type:text;"`
	IsTaxable           bool   `gorm:"type:text;"`
	Description         string          `gorm:"type:text;"`
}

// ServiceCategory represents the service_categories table
type ServiceCategory struct {
    ID         uint    `json:"ID"           gorm:"primaryKey;"`
    Name       string  `json:"Name"         gorm:"type:text;"`
    ProCost    float64 `json:"pro_cost"     gorm:"column:pro_cost;default:0"`
    RetailCost float64 `json:"retail_cost"  gorm:"column:retail_cost;default:0"`
}

// Int64Array handles Postgres integer[] scanning/valuing
type Int64Array []int64

func (a *Int64Array) Scan(value interface{}) error {
	if value == nil {
		*a = Int64Array{}
		return nil
	}
	var s string
	switch v := value.(type) {
	case []byte:
		s = string(v)
	case string:
		s = v
	default:
		return fmt.Errorf("Int64Array: unsupported scan type %T", value)
	}
	s = strings.TrimSpace(s)
	if s == "" || s == "{}" {
		*a = Int64Array{}
		return nil
	}
	if s[0] == '{' && s[len(s)-1] == '}' {
		s = s[1 : len(s)-1]
	}
	if strings.TrimSpace(s) == "" {
		*a = Int64Array{}
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]int64, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			return fmt.Errorf("Int64Array: parse %q: %w", p, err)
		}
		out = append(out, n)
	}
	*a = Int64Array(out)
	return nil
}

func (a Int64Array) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	parts := make([]string, len(a))
	for i, v := range a {
		parts[i] = strconv.FormatInt(v, 10)
	}
	return "{" + strings.Join(parts, ",") + "}", nil
}

// ServiceGroup represents the service_groups table
type ServiceGroup struct {
	ID         uint       `gorm:"primaryKey;" json:"ID"`
	Name       string     `gorm:"type:text;not null;index" json:"Name"`
	// Services   []Service  `gorm:"many2many:service_group_services;" json:"services,omitempty"`
	ServiceIDs Int64Array `gorm:"type:integer[];" json:"service_ids"`
	Description string     `gorm:"type:text;" json:"description"`
}

// ServiceType represents the service_types table
// type ServiceType struct {
// 	ID   uint   `gorm:"primaryKey;"`
// 	Name string `gorm:"type:text;not null;index"`
// }

// ServicePackage represents the service_packages table
// type ServicePackage struct {
// 	ID                 uint               `gorm:"primaryKey"`
// 	Name               string             `gorm:"type:text;not null;index"`
// 	Materials          []MaterialTemplate `gorm:"many2many:service_package_materials;"`
// 	Services           []Service          `gorm:"many2many:service_package_services;"`
// 	ResidentialEntryID uint
// 	ResidentialEntry   ManagerResidentialEntry `gorm:"foreignKey:ResidentialEntryID;constraint:OnDelete:SET NULL"`
// 	ResidentialTypeID  uint
// 	ResidentialType    ManagerResidentialType `gorm:"foreignKey:ResidentialTypeID;constraint:OnDelete:SET NULL"`
// 	ServiceGroups      []ServiceGroup         `gorm:"many2many:service_package_service_groups;"`
// 	IsActive           bool                   `json:"isactive" gorm:"default:false"`
// }

func (sd *ServiceDetails) Scan(value interface{}) error {
    bytes, ok := value.([]byte)
    if !ok {
        return fmt.Errorf("failed to scan ServiceDetails: expected []byte, got %T", value)
    }
    return json.Unmarshal(bytes, sd)
}

func (sd ServiceDetails) Value() (driver.Value, error) {
	return json.Marshal(sd)
}
