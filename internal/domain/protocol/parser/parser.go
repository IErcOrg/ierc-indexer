package parser

import (
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/IErcOrg/IERC_Indexer/internal/domain"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
)

type Parser interface {
	CheckFormat(data []byte) error
	Parse(tx *domain.Transaction) (protocol.IERCTransaction, error)
}

type parser struct {
	header       string
	headerLength int
	parsers      map[protocol.Protocol]Parser
}

func (p *parser) CheckFormat(data []byte) error {
	if len(data) == 0 {
		return protocol.NewProtocolError(protocol.NotProtocolData, "not protocol data")
	}

	dataStr := string(data)
	if !utf8.ValidString(dataStr) || !strings.HasPrefix(dataStr, p.header) {
		return protocol.NewProtocolError(protocol.NotProtocolData, "not protocol data")
	}

	var base struct {
		Protocol string `json:"p"`
		Operate  string `json:"op"`
	}

	err := json.Unmarshal([]byte(dataStr[len(p.header):]), &base)
	if err != nil {
		return protocol.NewProtocolError(protocol.InvalidProtocolFormat, "invalid protocol format")
	}

	if len(base.Protocol) == 0 {
		return protocol.NewProtocolError(protocol.UnknownProtocol, "unknown protocol")
	}

	return nil
}

func (p *parser) Parse(tx *domain.Transaction) (protocol.IERCTransaction, error) {

	var base struct {
		Protocol protocol.Protocol `json:"p"`
		Operate  protocol.Operate  `json:"op"`
	}

	if err := json.Unmarshal([]byte(tx.TxData[p.headerLength:]), &base); err != nil {
		return nil, protocol.NewProtocolError(protocol.InvalidProtocolFormat, "invalid protocol format")
	}

	parser, existed := p.parsers[base.Protocol]
	if !existed {
		return nil, protocol.NewProtocolError(protocol.UnknownProtocol, "unknown protocol")
	}

	return parser.Parse(tx)
}

func NewParser() Parser {

	parsers := make(map[protocol.Protocol]Parser)

	ierc20Parser := NewIERC20Parser(protocol.ProtocolHeader, protocol.TickETHI)
	parsers[protocol.ProtocolTERC20] = ierc20Parser
	parsers[protocol.ProtocolIERC20] = ierc20Parser
	parsers[protocol.ProtocolIERCPoW] = newIERC20PoWParser(protocol.ProtocolHeader)

	return &parser{
		header:       protocol.ProtocolHeader,
		headerLength: len(protocol.ProtocolHeader),
		parsers:      parsers,
	}
}
