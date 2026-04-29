package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"donapresentes/models"
)

const (
	XBZBaseURL = "https://api.minhaxbz.com.br:5001"
	XBZTimeout = 120 * time.Second
)

type XBZService struct {
	client *http.Client
	cnpj   string
	token  string
}

func NewXBZService(cnpj, token string) *XBZService {
	client := &http.Client{
		Timeout: XBZTimeout,
	}

	if cnpj == "" || token == "" {
		log.Printf("[XBZ] Aviso: CNPJ ou Token vazios")
	}

	return &XBZService{
		client: client,
		cnpj:   cnpj,
		token:  token,
	}
}

type XBZProduto struct {
	IdPessoa                             int        `json:"IdPessoa"`
	IdProduto                            int        `json:"IdProduto"`
	CodigoXbz                            string     `json:"CodigoXbz"`
	CodigoComposto                       string     `json:"CodigoComposto"`
	CodigoAmigavel                       string     `json:"CodigoAmigavel"`
	Nome                                 string     `json:"Nome"`
	Descricao                            string     `json:"Descricao"`
	SiteLink                             string     `json:"SiteLink"`
	ImageLink                            string     `json:"ImageLink"`
	WebTipoId                            int        `json:"WebTipoId"`
	WebTipo                              string     `json:"WebTipo"`
	WebSubTipoId                         int        `json:"WebSubTipoId"`
	WebSubTipo                           string     `json:"WebSubTipo"`
	CorWebPrincipalId                    int        `json:"CorWebPrincipalId"`
	CorWebPrincipal                      string     `json:"CorWebPrincipal"`
	CorWebSecundariaId                   int        `json:"CorWebSecundariaId"`
	CorWebSecundaria                     string     `json:"CorWebSecundaria"`
	Peso                                 float64    `json:"Peso"`
	Altura                               float64    `json:"Altura"`
	Largura                              float64    `json:"Largura"`
	Profundidade                         float64    `json:"Profundidade"`
	PrecoVenda                           float64    `json:"PrecoVenda"`
	PrecoVendaFormatado                  string     `json:"PrecoVendaFormatado"`
	PontaDeEstoque                       bool       `json:"PontaDeEstoque"`
	QuantidadeDisponivelEstoquePrincipal int        `json:"QuantidadeDisponivelEstoquePrincipal"`
	QuantidadeDisponivel                 int        `json:"QuantidadeDisponivel"`
	IdStatusConfiabilidade               int        `json:"IdStatusConfiabilidade"`
	StatusConfiabilidade                 string     `json:"StatusConfiabilidade"`
	Ncm                                  string     `json:"Ncm"`
	ReposicaoDataPrevista                *time.Time `json:"ReposicaoDataPrevista,omitempty"`
}

func (s *XBZService) GetProducts() ([]interface{}, error) {
	endpoint := fmt.Sprintf("%s/api/clientes/GetListaDeProdutos", XBZBaseURL)

	params := url.Values{}
	params.Add("cnpj", s.cnpj)
	params.Add("token", s.token)

	fullURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	body, err := s.fetch(fullURL)
	if err != nil {
		return nil, err
	}

	if typed, ok := s.parseTyped(body); ok {
		return toInterfaceSliceFromTyped(typed), nil
	}
	if gen, ok := s.parseGeneric(body); ok {
		return toInterfaceSliceFromGeneric(gen), nil
	}

	if inner, ok := s.parseQuoted(body); ok {
		if typed, ok := s.parseTyped(inner); ok {
			return toInterfaceSliceFromTyped(typed), nil
		}
		if gen, ok := s.parseGeneric(inner); ok {
			return toInterfaceSliceFromGeneric(gen), nil
		}
	}

	if fromFields, ok := s.parseFromFields(body); ok {
		return toInterfaceSliceFromGeneric(fromFields), nil
	}

	log.Printf("[XBZ] erro ao decodificar resposta da API (payload_size=%d)", len(body))
	return nil, fmt.Errorf("erro ao decodificar resposta da XBZ")
}

func (s *XBZService) fetch(fullURL string) ([]byte, error) {
	start := time.Now()
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("erro ao criar request: %w", err)
	}
	req.Header.Set("User-Agent", "DonaPresentes/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		duration := time.Since(start)
		return nil, fmt.Errorf("erro ao fazer requisição: %w (duracao=%s)", err, duration)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API retornou status %d: %s", resp.StatusCode, string(b))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		duration := time.Since(start)
		return nil, fmt.Errorf("erro ao ler resposta: %w (duracao=%s)", err, duration)
	}
	duration := time.Since(start)
	log.Printf("[XBZ] request concluído (duracao=%s, status=%d, payload_size=%d)", duration, resp.StatusCode, len(body))
	return body, nil
}

func (s *XBZService) parseTyped(body []byte) ([]XBZProduto, bool) {
	var arr []XBZProduto
	if err := json.Unmarshal(body, &arr); err == nil {
		return arr, true
	}
	return nil, false
}

func (s *XBZService) parseGeneric(body []byte) ([]map[string]interface{}, bool) {
	var arr []map[string]interface{}
	if err := json.Unmarshal(body, &arr); err == nil {
		return arr, true
	}
	return nil, false
}

func (s *XBZService) parseQuoted(body []byte) ([]byte, bool) {
	var asString string
	if err := json.Unmarshal(body, &asString); err == nil {
		return []byte(asString), true
	}
	return nil, false
}

func (s *XBZService) parseFromFields(body []byte) ([]map[string]interface{}, bool) {
	var asMap map[string]json.RawMessage
	if err := json.Unmarshal(body, &asMap); err != nil {
		return nil, false
	}
	for _, raw := range asMap {
		var typed []XBZProduto
		if err := json.Unmarshal(raw, &typed); err == nil {
			out := make([]map[string]interface{}, len(typed))
			for i, t := range typed {
				m := map[string]interface{}{
					"IdPessoa":             t.IdPessoa,
					"IdProduto":            t.IdProduto,
					"CodigoXbz":            t.CodigoXbz,
					"CodigoComposto":       t.CodigoComposto,
					"CodigoAmigavel":       t.CodigoAmigavel,
					"Nome":                 t.Nome,
					"Descricao":            t.Descricao,
					"ImageLink":            t.ImageLink,
					"WebTipo":              t.WebTipo,
					"QuantidadeDisponivel": t.QuantidadeDisponivel,
					"Ncm":                  t.Ncm,
				}
				out[i] = m
			}
			return out, true
		}
		var gen []map[string]interface{}
		if err := json.Unmarshal(raw, &gen); err == nil {
			return gen, true
		}
	}
	return nil, false
}

func toInterfaceSliceFromTyped(items []XBZProduto) []interface{} {
	out := make([]interface{}, len(items))
	for i, v := range items {
		out[i] = v
	}
	return out
}

func toInterfaceSliceFromGeneric(items []map[string]interface{}) []interface{} {
	out := make([]interface{}, len(items))
	for i, v := range items {
		out[i] = v
	}
	return out
}

func buildPhotosArray(imageLink string) []string {
	if imageLink == "" {
		return []string{}
	}
	return []string{imageLink}
}

func (s *XBZService) MapToLocalProduct(item interface{}) *models.Product {

	if xbzProd, ok := item.(XBZProduto); ok {

		if xbzProd.CodigoXbz == "" || xbzProd.Nome == "" {
			log.Printf("[XBZ] produto ignorado: CodigoXbz ou Nome vazios")
			return nil
		}

		now := time.Now()
		prod := &models.Product{
			ProductName:    xbzProd.Nome,
			InternalCode:   xbzProd.CodigoXbz,
			ProductGroup:   xbzProd.WebTipo,
			Description:    xbzProd.Descricao,
			Photos:         buildPhotosArray(xbzProd.ImageLink),
			NCM:            xbzProd.Ncm,
			MaterialOrigin: "",
			Stock:             xbzProd.QuantidadeDisponivel,
			SellingPrice:      xbzProd.PrecoVenda,
			KitType:           "none",
			IsComposition:     false,
			MovesStock:        true,
			EnabledForInvoice: true,
			Source:         "xbz",
			ImportedAt:     &now,
			LastSyncedAt:   &now,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		return prod
	}

	if m, ok := item.(map[string]interface{}); ok {
		getString := func(key string) string {
			if v, found := m[key]; found && v != nil {
				switch t := v.(type) {
				case string:
					return t
				case float64:
					return fmt.Sprintf("%v", t)
				default:
					return fmt.Sprintf("%v", t)
				}
			}
			return ""
		}
		getInt := func(key string) int {
			if v, found := m[key]; found && v != nil {
				switch t := v.(type) {
				case float64:
					return int(t)
				case int:
					return t
				case int64:
					return int(t)
				case string:
					if i, err := strconv.Atoi(t); err == nil {
						return i
					}
				}
			}
			return 0
		}
		getFloat64 := func(key string) float64 {
			if v, found := m[key]; found && v != nil {
				switch t := v.(type) {
				case float64:
					return t
				case int:
					return float64(t)
				case int64:
					return float64(t)
				case string:
					if f, err := strconv.ParseFloat(t, 64); err == nil {
						return f
					}
				}
			}
			return 0.0
		}

		internalCode := getString("CodigoXbz")
		nome := getString("Nome")
		if internalCode == "" || nome == "" {
			log.Printf("[XBZ] produto ignorado: CodigoXbz ou Nome vazios")
			return nil
		}

		now := time.Now()
		prod := &models.Product{
			ProductName:    nome,
			InternalCode:   internalCode,
			ProductGroup:   getString("WebTipo"),
			Description:    getString("Descricao"),
			Photos:         buildPhotosArray(getString("ImageLink")),
			NCM:            getString("Ncm"),
			MaterialOrigin: "",
			Stock:             getInt("QuantidadeDisponivel"),
			SellingPrice:      getFloat64("PrecoVenda"),
			KitType:           "none",
			IsComposition:     false,
			MovesStock:        true,
			EnabledForInvoice: true,
			Source:         "xbz",
			ImportedAt:     &now,
			LastSyncedAt:   &now,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		return prod
	}

	return nil
}

func (s *XBZService) MapToSupplier(item interface{}) *models.Supplier {

	nome := "XBZ Brindes"
	return &models.Supplier{
		Name:      nome,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (s *XBZService) GetXBZSupplierName() string {
	return "XBZ Brindes"
}
