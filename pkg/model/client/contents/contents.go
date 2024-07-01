package contents

import (
	"encoding/json"
	"fmt"

	"github.com/dmad1989/gophKeeper/pkg/model/enum"
	pb "github.com/dmad1989/gophKeeper/pkg/proto/gen"
)

type ItemFormatter interface {
	Format(description string) string
	Type() enum.ContentType
}

type Item struct {
	Content ItemFormatter
	Meta    string
}

func NewItem(c *pb.Content) (*Item, error) {
	switch enum.ContentType(c.Type) {
	case enum.LoginPassword:
		var loginPassword LoginPassword
		if err := json.Unmarshal(c.Data, &loginPassword); err != nil {
			return nil, fmt.Errorf("Contents.NewItem: loginPassword: json.Unmarshal: %w", err)
		}

		return &Item{Content: &loginPassword, Meta: c.Meta}, nil

	case enum.BankCard:
		var bankCard BankCard
		if err := json.Unmarshal(c.Data, &bankCard); err != nil {
			return nil, fmt.Errorf("Contents.NewItem: bankCard: json.Unmarshal: %w", err)
		}
		return &Item{Content: &bankCard, Meta: c.Meta}, nil
	}
	return nil, fmt.Errorf("Contents.NewItem: undefined type %v", c.Type)
}

func (i *Item) Format() string {
	return i.Content.Format(i.Meta)
}
