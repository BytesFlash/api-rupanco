package data

import (
	"fmt"
	"strings"

	"github.com/imedcl/manager-api/pkg/services"
)

type CGIUser struct {
	DNI          string    `json:"dni"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	PasswordDate string    `json:"password_date"`
	FlagDec      string    `json:"flag_dec"`
	Name         string    `json:"name"`
	Roles        []CGIRole `json:"roles"`
}

type CGIRole struct {
	Name           string `json:"name"`
	From           string `json:"from"`
	To             string `json:"to"`
	Place          string `json:"place"`
	System         string `json:"system"`
	Country        string `json:"country"`
	Institution    string `json:"institution"`
	InstitutionDNI string `json:"institution_dni"`
}

func GetCGIRoles(country string, institution string, offset int, limit int) ([]CGIUser, int, error) {
	var users []CGIUser
	var user CGIUser
	lastUserDNI := ""
	newOffset := offset
	for len(users) < limit {
		roles := services.ListUsersRoles(country, institution, newOffset, limit)
		if roles.Rows > 0 {
			for _, role := range roles.List {
				var autentiaRole CGIRole
				rut := strings.TrimSpace(role.Rut)
				if lastUserDNI != rut {
					if lastUserDNI != "" {
						users = append(users, user)
						user = CGIUser{}
						if len(users) >= limit {
							return users, newOffset, nil
						}
					}
					lastUserDNI = rut
					user.DNI = rut
					user.Email = role.Email
					user.Phone = role.Phone
					user.PasswordDate = role.PasswordDate
					user.FlagDec = role.FlagDec
					user.Name = role.Name
				}
				autentiaRole.Name = role.Role
				autentiaRole.From = role.RoleFrom
				autentiaRole.To = role.RoleTo
				autentiaRole.Country = role.Country
				autentiaRole.Institution = role.Institution
				autentiaRole.InstitutionDNI = role.RutInstitution
				autentiaRole.Place = role.Place
				autentiaRole.System = role.System
				user.Roles = append(user.Roles, autentiaRole)
				newOffset++
			}
		} else {
			break
		}
	}
	if user.DNI != "" {
		users = append(users, user)
	}
	return users, newOffset, nil
}

func GetCGIRole(dni string, country string, institution string) (CGIUser, error) {
	var user CGIUser
	roles := services.ListUserRoles(dni, country, institution)
	for _, role := range *roles {
		var autentiaRole CGIRole
		rut := strings.TrimSpace(role.Rut)
		user.DNI = rut
		user.Email = role.Email
		user.Phone = role.Phone
		user.PasswordDate = role.PasswordDate
		user.FlagDec = role.FlagDec
		user.Name = role.Name
		autentiaRole.Name = role.Role
		autentiaRole.From = role.RoleFrom
		autentiaRole.To = role.RoleTo
		autentiaRole.Country = role.Country
		autentiaRole.Institution = role.Institution
		autentiaRole.InstitutionDNI = role.RutInstitution
		autentiaRole.Place = role.Place
		autentiaRole.System = role.System
		user.Roles = append(user.Roles, autentiaRole)
	}
	return user, nil
}

func AddCGIRole(dni string, name string, country string, institution string) (bool, error) {
	response := services.CreateRole(dni, name, country, institution)
	return response, nil
}

func RemoveCGIRole(dni string, name string, country string, institution string) (bool, error) {
	response := services.DeleteRole(dni, name, country, institution)
	fmt.Println("REMOVE CGI ROLE FUNC:", response)
	return response, nil
}
