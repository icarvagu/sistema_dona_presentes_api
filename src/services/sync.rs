use crate::error::AppError;
use sqlx::PgPool;

pub struct SyncService;

impl SyncService {
    pub async fn sync_xbz_to_local(
        pool: &PgPool,
        xbz: &super::xbz::XBZService,
    ) -> Result<String, AppError> {
        let products = xbz.get_products().await?;
        let supplier_id = ensure_xbz_supplier(pool).await?;
        let mut imported = 0;
        let mut errors = 0;

        for p in &products {
            if should_skip_product(&p.codigo_xbz, &p.nome) {
                continue;
            }
            let supplier_code = p.codigo_amigavel.as_deref().unwrap_or("");
            let color = p.cor.as_deref().unwrap_or("");

            let result = sqlx::query(
                "INSERT INTO products (product_name, internal_code, supplier_id, supplier_code, product_group,
                        description, photos, ncm, stock, cost_price, source, imported_at, last_synced_at, color)
                 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'xbz',NOW(),NOW(),$11)
                 ON CONFLICT (internal_code, color) DO UPDATE SET
                        product_name = EXCLUDED.product_name,
                        supplier_id = EXCLUDED.supplier_id,
                        supplier_code = EXCLUDED.supplier_code,
                        product_group = EXCLUDED.product_group,
                        description = EXCLUDED.description,
                        photos = EXCLUDED.photos,
                        ncm = EXCLUDED.ncm,
                        stock = EXCLUDED.stock,
                        cost_price = EXCLUDED.cost_price,
                        last_synced_at = NOW(),
                        color = EXCLUDED.color",
            )
            .bind(&p.nome)
            .bind(&p.codigo_xbz)
            .bind(supplier_id)
            .bind(supplier_code)
            .bind(&p.web_tipo)
            .bind(&p.descricao)
            .bind(&p.image_link.as_ref().map(|u| vec![u.clone()]).unwrap_or_default())
            .bind(&p.ncm)
            .bind(p.quantidade_disponivel.unwrap_or(0))
            .bind(p.preco_venda.unwrap_or(0.0))
            .bind(color)
            .execute(pool)
            .await;

            match result {
                Ok(_) => imported += 1,
                Err(_) => errors += 1,
            }
        }

        Ok(format_sync_summary(imported, errors))
    }
}

async fn ensure_xbz_supplier(pool: &PgPool) -> Result<i32, AppError> {
    let supplier = sqlx::query_scalar::<_, i32>(
        "INSERT INTO suppliers (name, cnpj)
         VALUES ('XBZ', '36168035000181')
         ON CONFLICT (cnpj) WHERE cnpj IS NOT NULL DO UPDATE SET
            name = EXCLUDED.name,
            updated_at = NOW()
         RETURNING id",
    )
    .fetch_one(pool)
    .await
    .map_err(|e| AppError::internal(format!("XBZ supplier: {e}")))?;

    Ok(supplier)
}

fn should_skip_product(codigo_xbz: &str, nome: &str) -> bool {
    codigo_xbz.trim().is_empty() || nome.trim().is_empty()
}

fn format_sync_summary(imported: i32, errors: i32) -> String {
    format!("XBZ sync: {imported} updated, {errors} errors")
}

#[cfg(test)]
mod tests {
    use super::{format_sync_summary, should_skip_product};

    #[test]
    fn should_skip_product_when_code_missing() {
        assert!(should_skip_product("", "Nome"));
        assert!(should_skip_product("   ", "Nome"));
    }

    #[test]
    fn should_skip_product_when_name_missing() {
        assert!(should_skip_product("DN1", ""));
        assert!(should_skip_product("DN1", "   "));
    }

    #[test]
    fn should_not_skip_valid_product() {
        assert!(!should_skip_product("DN1", "Caneca"));
    }

    #[test]
    fn format_sync_summary_is_stable() {
        assert_eq!(format_sync_summary(12, 3), "XBZ sync: 12 updated, 3 errors");
    }
}
