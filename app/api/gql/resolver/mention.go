package resolver

import (
	"strings"

	"github.com/graph-gophers/graphql-go"
)

type Mention struct {
	Offset      int32
	Limit       int32
	Mentionable Mentionable
}

func ParseMentions(input string) (m []Mention) {
	chunks := strings.Split(input, " ")
	for _, chunk := range chunks {
		if len(chunk) == 0 {
			continue
		}
		var id graphql.ID
		err := id.UnmarshalGraphQL(chunk[1:])
		if err != nil {
			continue
		}
		offset := int32(strings.Index(input, chunk))
		limit := int32(len(chunk))
		if chunk[0] == '@' {
			m = append(m, Mention{
				Limit:  limit,
				Offset: offset,
				Mentionable: Mentionable{
					Type: "User",
					ID:   id,
				},
			})
		} else if chunk[0] == '#' {
			m = append(m, Mention{
				Limit:  limit,
				Offset: offset,
				Mentionable: Mentionable{
					Type: "Task",
					ID:   id,
				},
			})
		}
	}
	return
}

type Mentionable struct {
	// dep  *Dependencies
	Type string
	ID   graphql.ID
}

func (m Mentionable) ToUser() (*User, bool) {
	// fmt.Println("Mentionable: m.dep", m.dep)
	// if m.Type != "User" {
	// 	return nil, false
	// }
	// u, err := m.dep.Data.Get(m.ID)
	// if err != nil {
	// 	return nil, false
	// }
	// return &u, true
	return nil, false
}
