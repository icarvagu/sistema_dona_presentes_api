ALTER TABLE quotes
    ADD COLUMN IF NOT EXISTS freight_cnpj_solicitante  TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS freight_cnpj_cpf_origem   TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS freight_cnpj_cpf_destino  TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS freight_cnpj_devedor       TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS freight_tipo_transporte    TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS freight_contato            TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS freight_cidade_origem      TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS freight_cidade_destino     TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS freight_material           TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS freight_tipo_frete         TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS freight_produto            TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS freight_tipo_embalagem     TEXT         NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS freight_quantidade         INTEGER      NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS freight_volumes            JSONB        NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS freight_valor_nota         NUMERIC(12,2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS freight_peso_real          NUMERIC(10,3) NOT NULL DEFAULT 0;

