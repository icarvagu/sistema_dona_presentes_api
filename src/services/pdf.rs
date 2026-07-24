use crate::error::AppError;

pub async fn generate_order_pdf(
    pool: &sqlx::PgPool,
    sale_id: i32,
) -> Result<Vec<u8>, AppError> {
    let sale = crate::repositories::sale::get_by_id(pool, sale_id).await?;

    let mut items_text = String::new();
    for (i, item) in sale.items.iter().enumerate() {
        items_text.push_str(&format!(
            "{}. Prod #{} | Qtd: {} | Unit: R${:.2} | Total: R${:.2}\n",
            i + 1, item.product_id, item.quantity, item.unit_price, item.total_price
        ));
    }

    let content = format!(
        "PEDIDO #{}\n\
         ===========\n\
         Status: {}\n\
         Data: {}\n\
         Cliente ID: {}\n\
         Vendedor ID: {}\n\
         \n\
         ITENS:\n\
         {}\n\
         TOTAL: R${:.2}\n",
        sale_id,
        sale.status,
        sale.created_at.format("%d/%m/%Y"),
        sale.customer_id,
        sale.seller_id,
        items_text,
        sale.total_value
    );

    generate_simple_pdf(&format!("Pedido #{sale_id}"), &content)
}

pub async fn generate_quote_pdf(
    pool: &sqlx::PgPool,
    quote_id: i32,
) -> Result<Vec<u8>, AppError> {
    let quote = crate::repositories::quote::get_by_id(pool, quote_id).await?;

    let mut items_text = String::new();
    for (i, item) in quote.items.iter().enumerate() {
        items_text.push_str(&format!(
            "{}. Prod #{} | Qtd: {} | Unit: R${:.2} | Total: R${:.2}\n",
            i + 1, item.product_id, item.quantity, item.unit_price, item.total_price
        ));
    }

    let content = format!(
        "ORCAMENTO #{}\n\
         =============\n\
         Data: {}\n\
         Cliente ID: {}\n\
         Responsavel: {}\n\
         \n\
         ITENS:\n\
         {}\n\
         TOTAL: R${:.2}\n",
        quote_id,
        quote.created_at.format("%d/%m/%Y"),
        quote.customer_id,
        quote.responsible_name,
        items_text,
        quote.total_value
    );

    generate_simple_pdf(&format!("Orcamento #{quote_id}"), &content)
}

fn generate_simple_pdf(_title: &str, content: &str) -> Result<Vec<u8>, AppError> {
    let lines: Vec<&str> = content.lines().collect();
    let _line_count = lines.len();
    let line_height = 14.0;
    let base_y = 780.0;

    let mut stream = String::new();
    stream.push_str("BT\n");
    stream.push_str("/F1 10 Tf\n");

    for (i, line) in lines.iter().enumerate() {
        let escaped = line.replace('\\', "\\\\").replace('(', "\\(").replace(')', "\\)");
        let y = base_y - (i as f64 * line_height);
        stream.push_str(&format!("1 0 0 1 50 {} Tm\n", y));
        stream.push_str(&format!("({}) Tj\n", escaped));
    }

    stream.push_str("ET\n");
    let stream_bytes = stream.as_bytes();
    let stream_len = stream_bytes.len();

    let font_obj = "5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n";
    let content_obj = format!(
        "4 0 obj\n<< /Length {} >>\nstream\n{stream}\nendstream\nendobj\n",
        stream_len
    );
    let page_obj = format!(
        "3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792]\n\
         /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\nendobj\n"
    );
    let pages_obj = format!(
        "2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n"
    );
    let catalog_obj = "1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n";

    let header = "%PDF-1.4\n";
    let mut body = String::new();
    body.push_str(catalog_obj);
    body.push_str(&pages_obj);
    body.push_str(&page_obj);
    body.push_str(&content_obj);
    body.push_str(font_obj);

    let mut offsets = Vec::new();
    let mut pos = header.len() as i64;
    for line in body.lines() {
        offsets.push(pos);
        pos += line.len() as i64 + 1;
    }

    let xref_offset = pos;
    let obj_count = offsets.len() + 1;
    let mut xref = format!("xref\n0 {}\n0000000000 65535 f \n", obj_count);
    for offset in &offsets {
        xref.push_str(&format!("{:010} 00000 n \n", offset));
    }

    let trailer = format!(
        "trailer\n<< /Size {} /Root 1 0 R >>\nstartxref\n{xref_offset}\n%%EOF\n",
        obj_count
    );

    let mut result = Vec::new();
    result.extend_from_slice(header.as_bytes());
    result.extend_from_slice(body.as_bytes());
    result.extend_from_slice(xref.as_bytes());
    result.extend_from_slice(trailer.as_bytes());

    Ok(result)
}

#[cfg(test)]
mod tests {
    use super::generate_simple_pdf;

    #[test]
    fn generate_simple_pdf_returns_valid_markers() {
        let pdf = generate_simple_pdf("Teste", "linha 1\nlinha 2").unwrap();
        let text = String::from_utf8_lossy(&pdf);

        assert!(text.starts_with("%PDF-1.4"));
        assert!(text.contains("xref\n0 "));
        assert!(text.contains("trailer\n<< /Size "));
        assert!(text.ends_with("%%EOF\n"));
    }

    #[test]
    fn generate_simple_pdf_escapes_special_characters() {
        let pdf = generate_simple_pdf("Teste", "(abc) \\ value").unwrap();
        let text = String::from_utf8_lossy(&pdf);

        assert!(text.contains("\\(abc\\) \\\\ value"));
    }
}
