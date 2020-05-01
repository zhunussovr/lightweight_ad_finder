package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"gopkg.in/ldap.v3"
)

const (
	ldapServer = "192.168.15.5:389"
	//ldapPort     = 389
	ldapBind     = "vagrant@uplift.local"
	ldapPassword = "vagrant"

	filterDN = "(&(objectClass=person)(memberOf:1.2.840.113556.1.4.1941:=CN=stuff,OU=uGroups,OU=DEMO01,DC=uplift,DC=local)(|(sAMAccountName={username})(mail={username})))"
	baseDN   = "OU=DEMO01,DC=uplift,DC=local"

	loginUsername = "vagrant"
	loginPassword = "vagrant"
)

type uData struct {
	FullName, Email, Phone string
}

// Employee to define structure of json
type Employee struct {
	AccountName string
	UserData    []uData
}

func main() {
	conn, err := connect()

	if err != nil {
		fmt.Printf("Failed to connect. %s", err)
		return
	}

	defer conn.Close()

	if err := list(conn); err != nil {
		fmt.Printf("%v", err)
		return
	}

	if err := auth(conn); err != nil {
		fmt.Printf("%v", err)
		return
	}
}

func connect() (*ldap.Conn, error) {
	conn, err := ldap.Dial("tcp", ldapServer)

	if err != nil {
		return nil, fmt.Errorf("Failed to connect. %s", err)
	}

	if err := conn.Bind(ldapBind, ldapPassword); err != nil {
		return nil, fmt.Errorf("Failed to bind. %s", err)
	}

	return conn, nil
}

func list(conn *ldap.Conn) error {
	result, err := conn.Search(ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter("*"),
		[]string{"dn", "sAMAccountName", "mail", "sn", "givenName", "telephoneNumber"},
		nil,
	))

	if err != nil {
		return fmt.Errorf("Failed to search users. %s", err)
	}

	for _, entry := range result.Entries {

		userlist := Employee{
			AccountName: entry.GetAttributeValue("sAMAccountName"),
			UserData: []uData{
				{
					FullName: entry.GetAttributeValue("givenName"),
					Email:    entry.GetAttributeValue("mail"),
					Phone:    entry.GetAttributeValue("telephoneNumber"),
				},
			},
		}

		file, _ := json.MarshalIndent(userlist, "", " ")
		_ = ioutil.WriteFile("users.json", file, 0644)
	}

	return nil
}

func auth(conn *ldap.Conn) error {
	result, err := conn.Search(ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter(loginUsername),
		[]string{"dn"},
		nil,
	))

	if err != nil {
		return fmt.Errorf("Failed to find user. %s", err)
	}

	if len(result.Entries) < 1 {
		return fmt.Errorf("User does not exist")
	}

	if len(result.Entries) > 1 {
		return fmt.Errorf("Too many entries returned")
	}

	if err := conn.Bind(result.Entries[0].DN, loginPassword); err != nil {
		fmt.Printf("Failed to auth. %s", err)
	} else {
		fmt.Printf("Authenticated successfuly!")
	}

	return nil
}

func filter(needle string) string {
	res := strings.Replace(
		filterDN,
		"{username}",
		needle,
		-1,
	)

	return res
}
