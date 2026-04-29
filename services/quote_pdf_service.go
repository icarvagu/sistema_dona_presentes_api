package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"donapresentes/models"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type QuotePDFService struct {
	template *template.Template
}

type quoteTemplateData struct {
	QuoteNumber        string
	CreatedAt          string
	SellerName         string
	ResponsibleName    string
	CustomerName       string
	CustomerEmail      string
	CustomerPhone      string
	CustomerDocument   string
	CustomerAddress    string
	QuoteValidUntil    string
	ProductionLeadTime string
	TotalValue         string
	Items              []quoteTemplateItem
}

type quoteTemplateItem struct {
	InternalCode       string
	ImageURL           string
	ProductName        string
	Description        string
	NCM                string
	Personalization    string
	Quantity           string
	UnitPrice          string
	TotalPrice         string
}

func NewQuotePDFService() (*QuotePDFService, error) {
	tmpl, err := template.ParseFiles("templates/orcamento.html")
	if err != nil {
		return nil, fmt.Errorf("carregar template de orçamento: %w", err)
	}
	return &QuotePDFService{template: tmpl}, nil
}

func (s *QuotePDFService) GenerateQuotePDF(quote *models.Quote) ([]byte, error) {
	data := s.buildTemplateData(quote)
	var buf bytes.Buffer
	if err := s.template.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("executar template de orçamento: %w", err)
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
				WithMarginTop(0.2).
				WithMarginBottom(0.2).
				WithMarginLeft(0.2).
				WithMarginRight(0.2).
				WithPaperWidth(8.27).
				WithPaperHeight(11.69).
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("gerar PDF de orçamento: %w", err)
	}
	return pdfBuf, nil
}

func (s *QuotePDFService) buildTemplateData(quote *models.Quote) quoteTemplateData {
	data := quoteTemplateData{
		QuoteNumber:        quote.QuoteNumber,
		CreatedAt:          quote.CreatedAt.Format("02/01/2006"),
		SellerName:         "-",
		ResponsibleName:    quote.ResponsibleName,
		CustomerName:       "-",
		CustomerEmail:      "-",
		CustomerPhone:      "-",
		CustomerDocument:   "-",
		CustomerAddress:    "-",
		QuoteValidUntil:    "-",
		ProductionLeadTime: quote.ProductionLeadTime,
		TotalValue:         formatMoney(quote.TotalValue),
	}

	if data.QuoteNumber == "" {
		data.QuoteNumber = fmt.Sprintf("COT-%06d", quote.ID)
	}
	if quote.Seller != nil && quote.Seller.FullName != "" {
		data.SellerName = quote.Seller.FullName
	}
	if quote.QuoteValidUntil != nil {
		data.QuoteValidUntil = quote.QuoteValidUntil.Format("02/01/2006")
	}
	if quote.Customer != nil {
		data.CustomerName = quote.Customer.Name
		if quote.Customer.Email != "" {
			data.CustomerEmail = quote.Customer.Email
		}
		if quote.Customer.MobilePhone != "" {
			data.CustomerPhone = quote.Customer.MobilePhone
		} else if quote.Customer.BusinessPhone != "" {
			data.CustomerPhone = quote.Customer.BusinessPhone
		}
		if quote.Customer.CPF != nil {
			data.CustomerDocument = *quote.Customer.CPF
		} else if quote.Customer.CNPJ != nil {
			data.CustomerDocument = *quote.Customer.CNPJ
		}
		if len(quote.Customer.Addresses) > 0 {
			data.CustomerAddress = quote.Customer.Addresses[0].AddressLine
		}
	}

	for _, item := range quote.Items {
		productName := "-"
		description := "-"
		internalCode := "-"
		ncm := "-"
		imageURL := ""
		if item.Product != nil {
			productName = item.Product.ProductName
			internalCode = item.Product.InternalCode
			description = strings.TrimSpace(item.Product.Description)
			if description == "" {
				description = "-"
			}
			ncm = item.Product.NCM
			if ncm == "" {
				ncm = "-"
			}
			if len(item.Product.Photos) > 0 {
				imageURL = item.Product.Photos[0]
			}
		}
		personalization := strings.TrimSpace(item.PersonalizationType)
		if personalization == "" {
			personalization = "-"
		}
		data.Items = append(data.Items, quoteTemplateItem{
			InternalCode:    internalCode,
			ImageURL:        imageURL,
			ProductName:     productName,
			Description:     description,
			NCM:             ncm,
			Personalization: personalization,
			Quantity:        fmt.Sprintf("%d", item.Quantity),
			UnitPrice:       formatMoney(item.UnitPrice),
			TotalPrice:      formatMoney(item.TotalPrice),
		})
	}

	return data
}