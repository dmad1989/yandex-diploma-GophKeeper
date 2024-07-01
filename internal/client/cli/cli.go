package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/dmad1989/gophKeeper/pkg/model"
	"github.com/dmad1989/gophKeeper/pkg/model/client/contents"
	"github.com/dmad1989/gophKeeper/pkg/model/consts"
	"github.com/dmad1989/gophKeeper/pkg/model/enum"
	"github.com/dmad1989/gophKeeper/pkg/model/errs"
	pb "github.com/dmad1989/gophKeeper/pkg/proto/gen"
	"go.uber.org/zap"
	"golang.org/x/term"
)

const (
	maxCapacity = 1024 * 1024
	helpMsg     = "" +
		"available commands:\n" +
		"	'clear' - to clear terminal\n" +
		"\n" +
		"	'login' - to login\n" +
		"	'register' - to register\n" +
		"\n" +
		"	's [type]' - save content, where 'type' is: lp - LoginPassword, fl - File, bc - BankCard\n" +
		"\n" +
		"	'u [id]' - update content\n" +
		"	'd [id]' - delete content by id\n" +
		"	'l [type]' - get content by type, where 'type' is: lp - LoginPassword, fl - File, bc - BankCard\n	or get all if type is empty\n" +
		"	'g [id]' - get loginPassword or BankCard by id\n" +
		"	'gf [id]' - get file by id\n"
)

type Auth interface {
	Register(ctx context.Context, username, password string) (*pb.TokenData, error)
	Login(ctx context.Context, username, password string) (*pb.TokenData, error)
}

type Content interface {
	GetByType(ctx context.Context, contype enum.ContentType) ([]*model.Content, error)
	Get(ctx context.Context, id int32) (*contents.Item, error)
	Update(ctx context.Context, id int32, contype enum.ContentType, data []byte, meta string) error
	Delete(ctx context.Context, id int32) error
	Save(ctx context.Context, conType enum.ContentType, data []byte, meta string) (int32, error)
	SaveFile(ctx context.Context, path, meta string) (int32, error)
	GetFile(ctx context.Context, id int32) (string, error)
}

type cli struct {
	log      *zap.SugaredLogger
	scanner  *bufio.Scanner
	auth     Auth
	content  Content
	commands map[string]func(ctx context.Context, args []string) (string, error)
}

func New(ctx context.Context, a Auth, c Content) *cli {
	l := ctx.Value(consts.LoggerCtxKey).(*zap.SugaredLogger).Named("CLI")

	buf := make([]byte, maxCapacity)
	s := bufio.NewScanner(os.Stdin)
	s.Buffer(buf, maxCapacity)

	cli := &cli{
		log:     l,
		auth:    a,
		content: c,
		scanner: s,
	}
	cli.initCommands()
	return cli
}

func (c *cli) initCommands() {
	c.commands = map[string]func(ctx context.Context, args []string) (string, error){
		"login":    c.handleLogin,
		"register": c.handleRegistration,
		"s":        c.handleSave,
		"u":        c.handleUpdate,
		"d":        c.handleDelete,
		"l":        c.handleList,
		"g":        c.handleGet,
		"gf":       c.handleGetFile,
		"clear":    c.handleClear,
		"help":     c.handleHelp,
	}
}

func (c *cli) Start(ctx context.Context) {
	commandCh := make(chan struct{})
Loop:
	for {
		go c.processCommands(ctx, commandCh)
		select {
		case <-commandCh:
		case <-ctx.Done():
			fmt.Println("exit")
			break Loop
		}
	}
	fmt.Print("program is closed\n")
}

func (c *cli) processCommands(ctx context.Context, commandHandled chan struct{}) {
	defer func() {
		commandHandled <- struct{}{}
	}()
	cmd := c.readString("")
	if len(cmd) == 0 {
		return
	}
	result, err := c.handle(ctx, cmd)
	if err != nil {
		fmt.Printf("error: %v\n", err.Error())
		return
	}
	fmt.Printf("%s\n", result)
}

func (c *cli) handle(ctx context.Context, input string) (string, error) {
	arr := strings.Split(input, " ")

	command := arr[0]
	args := arr[1:]

	if f, ok := c.commands[command]; ok {
		return f(ctx, args)
	}
	return "", fmt.Errorf("command '%s' is not supported, type 'help' to display available commands", command)

}

func (c *cli) readString(label string) string {
	if len(label) != 0 {
		fmt.Println(label)
	}
	fmt.Print("-> ")
	c.scanner.Scan()
	return c.scanner.Text()
}

func (c *cli) handleClear(ctx context.Context, args []string) (string, error) {
	fmt.Print("\033[H\033[2J")
	return "", nil
}

func (c *cli) handleHelp(ctx context.Context, args []string) (string, error) {
	fmt.Print(helpMsg)
	return "", nil
}

func (c *cli) handleLogin(ctx context.Context, args []string) (string, error) {
	login := c.readString("input username")
	if login == "" {
		return "", errs.ErrInputLogin
	}
	password, err := c.readPassword()
	if err != nil {
		return "", fmt.Errorf("cli.handleLogin: %w", err)
	}
	_, err = c.auth.Login(ctx, login, password)
	if err != nil {
		return "", fmt.Errorf("cli.handleLogin: %w", err)
	}
	return "success", nil
}

func (c *cli) handleRegistration(ctx context.Context, args []string) (string, error) {
	login := c.readString("input username")
	if login == "" {
		return "", fmt.Errorf("cli.handleRegistration: %w", errs.ErrInputLogin)
	}
	password, err := c.readPassword()
	if err != nil {
		return "", fmt.Errorf("cli.handleRegistration: %w", err)
	}
	_, err = c.auth.Register(ctx, login, password)
	if err != nil {
		return "", fmt.Errorf("cli.handleRegistration: %w", err)
	}
	return "success", nil
}

func (c *cli) handleGet(ctx context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return "", errs.ErrEmptyArgID
	}
	id64, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		return "", fmt.Errorf("cli.handleGet: strconv.ParseInt: %w", err)
	}
	desc, err := c.content.Get(ctx, int32(id64))
	if err != nil {
		if errors.Is(err, errs.ErrContNotFound) {
			return "", errs.ErrContNotFound
		}
		return "", fmt.Errorf("cli.handleGet: %w", err)
	}

	return desc.Format(), nil
}

func (c *cli) handleList(ctx context.Context, args []string) (string, error) {
	cType := enum.Nan
	if len(args) != 0 {
		if aType, ok := enum.ArgToType[args[0]]; ok {
			cType = aType
		}
	}

	cs, err := c.content.GetByType(ctx, cType)
	if err != nil {
		if errors.Is(err, errs.ErrContNotFound) {
			return "", errs.ErrContNotFound
		}
		return "", fmt.Errorf("cli.handleList: %w", err)
	}
	var writer strings.Builder
	if len(cs) == 0 {
		_, err := writer.WriteString("empty")
		if err != nil {
			return "", fmt.Errorf("cli.handleList: writer.WriteString(\"empty\") %w", err)
		}
	}
	for _, c := range cs {
		_, err := writer.WriteString(fmt.Sprintf("id: %d - type: '%s', descr: '%s'\n", c.ID, enum.TypeToArg[c.Type], c.Meta))
		if err != nil {
			return "", fmt.Errorf("cli.handleList: writer.WriteString(fmt.Sprintf) %w", err)
		}
	}
	return writer.String(), nil
}

func (c *cli) handleSave(ctx context.Context, args []string) (res string, err error) {
	if len(args) == 0 {
		err = fmt.Errorf("cli.handleSave: %w", errs.ErrEmptyArgType)
		return
	}
	var content any
	var meta string
	ctype := args[0]
	switch ctype {
	case consts.LoginPassword:
		content, meta, err = c.readLoginPassword()
		if err != nil {
			err = fmt.Errorf("cli.handleSave: %w", err)
			return
		}
		res, err = c.saveTextContent(ctx, content, meta, enum.LoginPassword)
	case consts.BankCard:
		content, meta, err = c.readBankCard()
		if err != nil {
			err = fmt.Errorf("cli.handleUpdate: %w", err)
			return
		}
		res, err = c.saveTextContent(ctx, content, meta, enum.BankCard)
	case consts.File:
		res, err = c.saveFile(ctx)
	default:
		err = fmt.Errorf("content type argument '%s' is not supported, type 'help' to display available types", ctype)
	}

	if err != nil {
		err = fmt.Errorf("cli.handleSave: %w", err)
		return
	}
	return
}

func (c *cli) handleUpdate(ctx context.Context, args []string) (res string, err error) {
	if len(args) == 0 {
		err = errs.ErrEmptyArgID
		return
	}
	id64, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		err = fmt.Errorf("cli.handleUpdate: strconv.ParseInt: %w", err)
		return
	}
	id := int32(id64)
	var content any
	cItem, err := c.content.Get(ctx, id)
	if err != nil {
		err = fmt.Errorf("cli.handleUpdate: %w", err)
		return
	}
	var meta string
	switch cItem.Content.Type() {
	case enum.LoginPassword:
		content, meta, err = c.readLoginPassword()
		if err != nil {
			err = fmt.Errorf("cli.handleUpdate: %w", err)
			return
		}
		res, err = c.updateTextContent(ctx, id, content, meta, enum.LoginPassword)
	case enum.BankCard:
		content, meta, err = c.readBankCard()
		if err != nil {
			err = fmt.Errorf("cli.handleUpdate: %w", err)
			return
		}
		res, err = c.updateTextContent(ctx, id, content, meta, enum.BankCard)
	case enum.File:
		err = errs.ErrFileUpdate
	default:
		err = fmt.Errorf("content type argument '%d' is not supported, type 'help' to display available types", cItem.Content.Type())
	}

	if err != nil {
		if errors.Is(err, errs.ErrContNotFound) {
			return "", errs.ErrContNotFound
		}
		err = fmt.Errorf("cli.handleUpdate: %w", err)
		return
	}
	return
}

func (c *cli) handleDelete(ctx context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return "", errs.ErrEmptyArgID
	}
	id64, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		return "", fmt.Errorf("cli.handleDelete: strconv.ParseInt: %w", err)
	}
	err = c.content.Delete(ctx, int32(id64))
	if err != nil {
		if errors.Is(err, errs.ErrContNotFound) {
			return "", errs.ErrContNotFound
		}
		return "", fmt.Errorf("cli.handleDelete: %w", err)
	}
	return "deleted", nil
}

func (c *cli) saveTextContent(ctx context.Context, content any, meta string, cType enum.ContentType) (string, error) {
	contJson, err := json.Marshal(content)
	if err != nil {
		return "", fmt.Errorf("cli.saveTextContent: json.Marshal: %w", err)
	}
	id, err := c.content.Save(ctx, cType, contJson, meta)
	if err != nil {
		return "", fmt.Errorf("cli.saveTextContent: %w", err)
	}
	return fmt.Sprintf("saved successfully, id: %v", id), nil
}

func (c *cli) updateTextContent(ctx context.Context, id int32, content any, meta string, cType enum.ContentType) (string, error) {
	contJson, err := json.Marshal(content)
	if err != nil {
		return "", fmt.Errorf("cli.updateTextContent: json.Marshal: %w", err)
	}
	err = c.content.Update(ctx, id, cType, contJson, meta)
	if err != nil {
		return "", fmt.Errorf("cli.updateTextContent: %w", err)
	}
	return fmt.Sprintf("updated successfully, id: %v", id), nil
}

func (c *cli) handleGetFile(ctx context.Context, args []string) (string, error) {
	if len(args) == 0 {
		return "", errs.ErrEmptyArgID
	}
	id64, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		return "", fmt.Errorf("cli.handleGetFile: strconv.ParseInt: %w", err)
	}

	path, err := c.content.GetFile(ctx, int32(id64))
	if err != nil {
		return "", fmt.Errorf("cli.handleGetFile: %w", err)
	}

	return fmt.Sprintf("recieved file saved to: %v", path), nil
}

func (c *cli) saveFile(ctx context.Context) (string, error) {
	filePath := c.readString("input file path")
	if filePath == "" {
		return "", errs.ErrInputFilePath
	}

	meta := c.readString("input description")
	if meta == "" {
		return "", errs.ErrInputDesc
	}

	id, err := c.content.SaveFile(ctx, filePath, meta)
	if err != nil {
		return "", fmt.Errorf("cli.saveFile: %w", err)
	}
	return fmt.Sprintf("%d", id), nil
}

func (c *cli) readLoginPassword() (lp *contents.LoginPassword, desc string, err error) {
	login := c.readString("input login")
	if login == "" {
		err = fmt.Errorf("readLoginPassword: %w", errs.ErrInputLogin)
		return
	}
	password, err := c.readPassword()
	if err != nil {
		err = fmt.Errorf("readLoginPassword: %w", err)
	}
	desc = c.readString("input description")
	if desc == "" {
		err = fmt.Errorf("readLoginPassword: %w", errs.ErrInputDesc)
	}
	lp = contents.NewLoginPassword(login, password)
	return
}

func (c *cli) readBankCard() (bc *contents.BankCard, desc string, err error) {
	number := c.readString("input number")
	if number == "" {
		err = fmt.Errorf("readBankCard: %w", errs.ErrInputBCNumber)
		return
	}
	expireAt := c.readString("input expireAt in format: MM/YY")
	if expireAt == "" {
		err = fmt.Errorf("readBankCard: %w", errs.ErrInputBCExpireDate)
		return
	}
	name := c.readString("input name")
	if name == "" {
		err = fmt.Errorf("readBankCard: %w", errs.ErrInputBCName)
		return
	}
	surname := c.readString("input surname")
	if surname == "" {
		err = fmt.Errorf("readBankCard: %w", errs.ErrInputBCSurname)
		return
	}
	desc = c.readString("input description")
	if desc == "" {
		err = fmt.Errorf("readBankCard: %w", errs.ErrInputDesc)
		return
	}

	bc = contents.NewBankCard(number, expireAt, name, surname)
	return
}

func (c *cli) readPassword() (string, error) {
	fmt.Println("password:")
	fmt.Print("-> ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		c.log.Errorw("readPassword: term.ReadPassword:", zap.Error(err))
		return "", errors.New("internal, try again")
	}
	if len(bytePassword) == 0 {
		return "", errs.ErrInputPassword
	}
	fmt.Println()
	return string(bytePassword), nil
}
