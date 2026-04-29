package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"
	"time"

	"donapresentes/models"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

const (
	companyName    = "DONNA BRINDES E PRESENTES LTDA"
	companyCNPJ    = "36.168.035/0001-81"
	companyAddress = "Avenida Deputado Emílio Carlos, 299 - Limão"
	companyCity    = "São Paulo/SP - CEP: 02721-000"
	companyPhones  = "(11)3966-2135 / (11) 91713-1001"
	companyEmail   = "vendas@donnapresentes.com.br"
	companyWebsite = "www.donnapresentes.com.br"
)

type PDFService struct {
	template *template.Template
}

type pedidoTemplateData struct {
	SaleID           int
	SaleIDFormatted  string
	LogoDataURI      template.URL
	SellerName       string
	DataPedido       string
	EmailFinanceiro  string
	IsPF             bool
	ClienteNome      string
	ClienteEndereco  string
	ClienteCidade    string
	ClienteTelefone  string
	ClienteDoc       string
	ClienteCEP       string
	ClienteEstado    string
	ClienteEmail     string
	ClienteNomeFantasia string
	Vencimento       string
	ValorTotal       string
	ValorProdutos    string
	ValorFrete       string
	FormaPagamento   string
	AosCuidadosDe    string
	Transportadora   string
	PrazoEntrega     string
	Observacoes      string
	Items            []pedidoItemData
}

type pedidoItemData struct {
	ProductName        string
	QuantityFormatted  string
	UnitPriceFormatted string
	TotalPriceFormatted string
	Logo               string
	Cores              string
	Grav               string
	Local              string
	InternalCode       string
	NCM                string
	CorItem            string
	Obs                string
}

func NewPDFService() (*PDFService, error) {
	tmpl, err := template.ParseFiles("templates/pedido.html")
	if err != nil {
		return nil, fmt.Errorf("carregar template: %w", err)
	}
	return &PDFService{template: tmpl}, nil
}

func (s *PDFService) GenerateOrderPDF(sale *models.Sale) ([]byte, error) {
	data := s.buildTemplateData(sale)
	var buf bytes.Buffer
	if err := s.template.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("executar template: %w", err)
	}
	html := buf.String()

	opts := chromedp.DefaultExecAllocatorOptions[:]
	if _, err := os.Stat("/usr/bin/chromium"); err == nil {
		opts = append(opts, chromedp.ExecPath("/usr/bin/chromium"))
	}
	opts = append(opts,
		chromedp.Flag("headless", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-software-rasterizer", true),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var pdfBuf []byte
	dataURL := "data:text/html;base64," + base64.StdEncoding.EncodeToString([]byte(html))

	err := chromedp.Run(ctx,
		chromedp.Navigate(dataURL),
		chromedp.WaitReady("body"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithMarginTop(0.01).
				WithMarginBottom(0.01).
				WithMarginLeft(0.01).
				WithMarginRight(0.01).
				WithPaperWidth(8.27).
				WithPaperHeight(11.69).
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("gerar PDF: %w", err)
	}
	return pdfBuf, nil
}

func (s *PDFService) buildTemplateData(sale *models.Sale) pedidoTemplateData {
	data := pedidoTemplateData{
		SaleID:          sale.ID,
		SaleIDFormatted: fmt.Sprintf("%04d", sale.ID),
		LogoDataURI:     loadLogoDataURI(),
		SellerName:      "-",
		DataPedido:      sale.CreatedAt.Format("02/01/2006"),
		EmailFinanceiro: "-",
		Vencimento:      "/  /",
		ValorTotal:      formatMoney(sale.TotalValue),
		ValorProdutos:   formatMoney(sale.TotalValue),
		ValorFrete:      "0,00",
		FormaPagamento:  sale.PaymentMethod,
		AosCuidadosDe:   "-",
		Transportadora:  "-",
		PrazoEntrega:    "-",
		Observacoes:     "-",
	}

	if sale.Seller != nil {
		data.SellerName = sale.Seller.FullName
	}
	if sale.FirstInstallmentStart != nil {
		data.Vencimento = sale.FirstInstallmentStart.Format("02/01/2006")
	}
	if len(sale.Carriers) > 0 {
		names := make([]string, len(sale.Carriers))
		for i, c := range sale.Carriers {
			names[i] = c.Name
		}
		data.Transportadora = strings.Join(names, ", ")
	}

	if sale.Customer != nil {
		c := sale.Customer
		data.IsPF = c.CustomerType == "PF"
		data.ClienteNome = c.Name
		data.ClienteEmail = c.Email
		data.ClienteTelefone = c.MobilePhone
		if data.ClienteTelefone == "" {
			data.ClienteTelefone = c.BusinessPhone
		}
		data.ClienteEndereco = "-"
		if len(c.Addresses) > 0 {
			data.ClienteEndereco = c.Addresses[0].AddressLine
		}
		data.ClienteCidade = "-"
		data.ClienteCEP = "-"
		data.ClienteEstado = "-"
		data.ClienteNomeFantasia = "-"
		if c.CustomerType == "PF" && c.CPF != nil {
			data.ClienteDoc = *c.CPF
		} else if c.CustomerType == "PJ" && c.CNPJ != nil {
			data.ClienteDoc = *c.CNPJ
		} else {
			data.ClienteDoc = "-"
		}
	}

	for _, item := range sale.Items {
		productName := "-"
		internalCode := "-"
		ncm := "-"
		if item.Product != nil {
			productName = item.Product.ProductName
			internalCode = item.Product.InternalCode
			ncm = item.Product.NCM
		}
		data.Items = append(data.Items, pedidoItemData{
			ProductName:        productName,
			QuantityFormatted:  formatQty(item.Quantity),
			UnitPriceFormatted: formatMoney(item.UnitPrice),
			TotalPriceFormatted: formatMoney(item.TotalPrice),
			Logo:               "-",
			Cores:              "-",
			Grav:               "-",
			Local:              "-",
			InternalCode:       internalCode,
			NCM:                ncm,
			CorItem:            "-",
			Obs:                "-",
		})
	}
	return data
}

func formatMoney(v float64) string {
	return strings.Replace(fmt.Sprintf("%.2f", v), ".", ",", 1)
}

func formatQty(qty int) string {
	if qty >= 1000 {
		return fmt.Sprintf("%d.%03d", qty/1000, qty%1000)
	}
	return strconv.Itoa(qty)
}

func loadLogoDataURI() template.URL {
	logoBytes, err := os.ReadFile("templates/img/logo-donna.png")
	if err != nil {
		return ""
	}
	return template.URL("data:image/png;base64," + base64.StdEncoding.EncodeToString(logoBytes))
}
