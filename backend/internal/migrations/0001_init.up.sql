CREATE TABLE public.accounts (
    account_id integer NOT NULL,
    account_code character varying(100) NOT NULL,
    account_name character varying(128) NOT NULL,
    account_type character varying(100) NOT NULL,
    parent_id integer,
    created_at timestamp without time zone DEFAULT now(),
    updated_at timestamp without time zone DEFAULT now()
);

CREATE SEQUENCE public.accounts_account_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.accounts_account_id_seq OWNED BY public.accounts.account_id;
CREATE TABLE public.ad_invoices (
    invoice_no character varying(100) NOT NULL,
    username character varying(100),
    store character varying(128),
    invoice_date date,
    total numeric,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);
CREATE TABLE public.asset_accounts (
    id integer NOT NULL,
    account_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);
CREATE SEQUENCE public.asset_accounts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.asset_accounts_id_seq OWNED BY public.asset_accounts.id;
CREATE TABLE public.batch_history (
    id integer NOT NULL,
    process_type text NOT NULL,
    started_at timestamp with time zone DEFAULT now() NOT NULL,
    total_data integer NOT NULL,
    done_data integer DEFAULT 0 NOT NULL,
    status text DEFAULT 'processing'::text NOT NULL,
    error_message text,
    file_name text,
    file_path text
);
CREATE TABLE public.batch_history_details (
    id integer NOT NULL,
    batch_id integer NOT NULL,
    reference text NOT NULL,
    store text,
    status text NOT NULL,
    error_message text
);
CREATE SEQUENCE public.batch_history_details_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.batch_history_details_id_seq OWNED BY public.batch_history_details.id;
CREATE SEQUENCE public.batch_history_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.batch_history_id_seq OWNED BY public.batch_history.id;
CREATE TABLE public.cached_metrics (
    id integer NOT NULL,
    shop_username character varying(100) NOT NULL,
    period character varying(100) NOT NULL,
    sum_revenue numeric(14,2) NOT NULL,
    sum_cogs numeric(14,2) NOT NULL,
    sum_fees numeric(14,2) NOT NULL,
    net_profit numeric(14,2) NOT NULL,
    ending_cash_balance numeric(14,2) NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);
CREATE SEQUENCE public.cached_metrics_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.cached_metrics_id_seq OWNED BY public.cached_metrics.id;
CREATE TABLE public.dropship_purchase_details (
    id integer NOT NULL,
    kode_pesanan text NOT NULL,
    sku text,
    nama_produk text,
    harga_produk numeric,
    qty integer,
    total_harga_produk numeric,
    harga_produk_channel numeric,
    total_harga_produk_channel numeric,
    potensi_keuntungan numeric
);
CREATE SEQUENCE public.dropship_purchase_details_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.dropship_purchase_details_id_seq OWNED BY public.dropship_purchase_details.id;
CREATE TABLE public.dropship_purchases (
    kode_pesanan text NOT NULL,
    kode_transaksi text,
    waktu_pesanan_terbuat timestamp without time zone,
    status_pesanan_terakhir text,
    biaya_lainnya numeric,
    biaya_mitra_jakmall numeric,
    total_transaksi numeric,
    dibuat_oleh text,
    jenis_channel text,
    nama_toko text,
    kode_invoice_channel text,
    gudang_pengiriman text,
    jenis_ekspedisi text,
    cashless text,
    nomor_resi text,
    waktu_pengiriman timestamp without time zone,
    provinsi text,
    kota text
);
CREATE TABLE public.expense_lines (
    line_id integer NOT NULL,
    expense_id uuid NOT NULL,
    account_id integer NOT NULL,
    amount numeric NOT NULL
);
CREATE SEQUENCE public.expense_lines_line_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.expense_lines_line_id_seq OWNED BY public.expense_lines.line_id;
CREATE TABLE public.expenses (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    date timestamp without time zone NOT NULL,
    description text NOT NULL,
    amount numeric NOT NULL,
    asset_account_id integer NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);
CREATE TABLE public.jenis_channels (
    jenis_channel_id integer NOT NULL,
    jenis_channel text NOT NULL
);
CREATE SEQUENCE public.jenis_channels_jenis_channel_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.jenis_channels_jenis_channel_id_seq OWNED BY public.jenis_channels.jenis_channel_id;
CREATE TABLE public.journal_entries (
    journal_id integer NOT NULL,
    entry_date date NOT NULL,
    description text,
    source_type character varying(100) NOT NULL,
    source_id character varying(100) NOT NULL,
    shop_username character varying(100) NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    store text DEFAULT ''::text NOT NULL
);
CREATE SEQUENCE public.journal_entries_journal_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.journal_entries_journal_id_seq OWNED BY public.journal_entries.journal_id;
CREATE TABLE public.journal_lines (
    line_id integer NOT NULL,
    journal_id integer NOT NULL,
    account_id integer NOT NULL,
    is_debit boolean NOT NULL,
    amount numeric(14,2) NOT NULL,
    memo text
);
CREATE SEQUENCE public.journal_lines_line_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.journal_lines_line_id_seq OWNED BY public.journal_lines.line_id;
CREATE TABLE public.reconciled_transactions (
    id integer NOT NULL,
    shop_username character varying(100) NOT NULL,
    dropship_id character varying(100),
    shopee_id character varying(100),
    status character varying(100) NOT NULL,
    matched_at timestamp without time zone DEFAULT now() NOT NULL
);
CREATE SEQUENCE public.reconciled_transactions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.reconciled_transactions_id_seq OWNED BY public.reconciled_transactions.id;
CREATE TABLE public.shopee_adjustments (
    id integer NOT NULL,
    nama_toko character varying(128) NOT NULL,
    tanggal_penyesuaian date NOT NULL,
    tipe_penyesuaian character varying(128) NOT NULL,
    alasan_penyesuaian text,
    biaya_penyesuaian numeric NOT NULL,
    no_pesanan character varying(100) NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);
CREATE SEQUENCE public.shopee_adjustments_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.shopee_adjustments_id_seq OWNED BY public.shopee_adjustments.id;
CREATE TABLE public.shopee_affiliate_sales (
    kode_pesanan character varying(100),
    status_pesanan character varying(100),
    status_terverifikasi character varying(100),
    waktu_pesanan timestamp without time zone,
    waktu_pesanan_selesai timestamp without time zone,
    waktu_pesanan_terverifikasi timestamp without time zone,
    kode_produk character varying(100),
    nama_produk text,
    id_model character varying(100),
    l1_kategori_global character varying(100),
    l2_kategori_global character varying(100),
    l3_kategori_global character varying(100),
    kode_promo character varying(100),
    harga numeric,
    jumlah integer,
    nama_affiliate character varying(128),
    username_affiliate character varying(128),
    mcn_terhubung character varying(128),
    id_komisi_pesanan character varying(100),
    partner_promo character varying(100),
    jenis_promo character varying(100),
    nilai_pembelian numeric,
    jumlah_pengembalian numeric,
    tipe_pesanan character varying(100),
    estimasi_komisi_per_produk numeric,
    estimasi_komisi_affiliate_per_produk numeric,
    persentase_komisi_affiliate_per_produk numeric,
    estimasi_komisi_mcn_per_produk numeric,
    persentase_komisi_mcn_per_produk numeric,
    estimasi_komisi_per_pesanan numeric,
    estimasi_komisi_affiliate_per_pesanan numeric,
    estimasi_komisi_mcn_per_pesanan numeric,
    catatan_produk text,
    platform character varying(100),
    tingkat_komisi numeric,
    pengeluaran numeric,
    status_pemotongan character varying(100),
    metode_pemotongan character varying(100),
    waktu_pemotongan timestamp without time zone,
    nama_toko text
);
CREATE TABLE public.shopee_order_details (
    order_sn character varying(64) NOT NULL,
    nama_toko text,
    created_at timestamp without time zone DEFAULT now(),
    status text,
    checkout_time timestamp without time zone,
    update_time timestamp without time zone,
    pay_time timestamp without time zone,
    total_amount numeric,
    currency text,
    actual_shipping_fee_confirmed boolean,
    buyer_cancel_reason text,
    buyer_cpf_id text,
    buyer_user_id bigint,
    buyer_username text,
    cancel_by text,
    cancel_reason text,
    cod boolean,
    create_time timestamp without time zone,
    days_to_ship integer,
    dropshipper text,
    dropshipper_phone text,
    estimated_shipping_fee numeric,
    fulfillment_flag text,
    goods_to_declare boolean,
    message_to_seller text,
    note text,
    note_update_time timestamp without time zone,
    order_status text,
    pickup_done_time timestamp without time zone,
    region text,
    reverse_shipping_fee numeric,
    ship_by_date timestamp without time zone,
    shipping_carrier text,
    split_up boolean,
    payment_method text,
    recipient_name text,
    recipient_phone text,
    recipient_full_address text,
    recipient_city text,
    recipient_district text,
    recipient_state text,
    recipient_town text,
    recipient_zipcode text
);
CREATE TABLE public.shopee_order_items (
    id integer NOT NULL,
    order_sn character varying(64),
    order_item_id bigint,
    item_name text,
    model_original_price numeric,
    model_quantity_purchased integer,
    item_id bigint,
    item_sku text,
    model_id bigint,
    model_name text,
    model_sku text,
    model_discounted_price numeric,
    weight numeric,
    promotion_id bigint,
    promotion_type text,
    promotion_group_id bigint,
    add_on_deal boolean,
    add_on_deal_id bigint,
    main_item boolean,
    is_b2c_owned_item boolean,
    is_prescription_item boolean,
    wholesale boolean,
    product_location_id text[],
    image_url text
);
CREATE SEQUENCE public.shopee_order_items_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.shopee_order_items_id_seq OWNED BY public.shopee_order_items.id;
CREATE TABLE public.shopee_order_packages (
    id integer NOT NULL,
    order_sn character varying(64),
    package_number text,
    logistics_status text,
    shipping_carrier text,
    logistics_channel_id bigint,
    parcel_chargeable_weight_gram integer,
    allow_self_design_awb boolean,
    sorting_group text,
    group_shipment_id text
);
CREATE SEQUENCE public.shopee_order_packages_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.shopee_order_packages_id_seq OWNED BY public.shopee_order_packages.id;
CREATE TABLE public.shopee_settled (
    no_pesanan character varying(100),
    no_pengajuan character varying(100),
    username_pembeli character varying(100),
    waktu_pesanan_dibuat date,
    metode_pembayaran_pembeli character varying(100),
    tanggal_dana_dilepaskan date,
    harga_asli_produk numeric,
    total_diskon_produk numeric,
    jumlah_pengembalian_dana_ke_pembeli numeric,
    diskon_produk_dari_shopee numeric,
    diskon_voucher_ditanggung_penjual numeric,
    cashback_koin_ditanggung_penjual numeric,
    ongkir_dibayar_pembeli numeric,
    diskon_ongkir_ditanggung_jasa_kirim numeric,
    gratis_ongkir_dari_shopee numeric,
    ongkir_yang_diteruskan_oleh_shopee_ke_jasa_kirim numeric,
    ongkos_kirim_pengembalian_barang numeric,
    pengembalian_biaya_kirim numeric,
    biaya_komisi_ams numeric,
    biaya_administrasi numeric,
    biaya_layanan_termasuk_ppn_11 numeric,
    premi numeric,
    biaya_program numeric,
    biaya_kartu_kredit numeric,
    biaya_kampanye numeric,
    bea_masuk_ppn_pph numeric,
    total_penghasilan numeric,
    kompensasi numeric,
    promo_gratis_ongkir_dari_penjual numeric,
    jasa_kirim character varying(100),
    nama_kurir character varying(100),
    pengembalian_dana_ke_pembeli numeric,
    pro_rata_koin_yang_ditukarkan_untuk_pengembalian_barang numeric,
    pro_rata_voucher_shopee_untuk_pengembalian_barang numeric,
    pro_rated_bank_payment_channel_promotion_for_returns numeric,
    pro_rated_shopee_payment_channel_promotion_for_returns numeric,
    nama_toko character varying(100),
    is_data_mismatch boolean DEFAULT false NOT NULL,
    is_settled_confirmed boolean DEFAULT false NOT NULL,
    biaya_transaksi numeric DEFAULT 0
);
CREATE TABLE public.shopee_settled_orders (
    id integer NOT NULL,
    order_id character varying(100) NOT NULL,
    net_income numeric(12,2) NOT NULL,
    service_fee numeric(12,2) NOT NULL,
    campaign_fee numeric(12,2) NOT NULL,
    credit_card_fee numeric(12,2) NOT NULL,
    shipping_subsidy numeric(12,2) NOT NULL,
    tax_and_import_fee numeric(12,2) NOT NULL,
    settled_date date NOT NULL,
    seller_username character varying(100) NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL
);
CREATE SEQUENCE public.shopee_settled_orders_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.shopee_settled_orders_id_seq OWNED BY public.shopee_settled_orders.id;
CREATE TABLE public.stores (
    store_id integer NOT NULL,
    jenis_channel_id integer NOT NULL,
    nama_toko text NOT NULL,
    code_id text,
    shop_id text,
    access_token text,
    refresh_token text,
    expire_in integer,
    request_id text,
    last_updated timestamp without time zone
);
CREATE SEQUENCE public.stores_store_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.stores_store_id_seq OWNED BY public.stores.store_id;
CREATE TABLE public.tax_payments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    store text NOT NULL,
    period_type text NOT NULL,
    period_value text NOT NULL,
    revenue numeric NOT NULL,
    tax_rate numeric DEFAULT 0.005 NOT NULL,
    tax_amount numeric NOT NULL,
    is_paid boolean DEFAULT false NOT NULL,
    paid_at timestamp without time zone
);
CREATE TABLE public.withdrawals (
    id integer NOT NULL,
    store character varying(128) NOT NULL,
    date date NOT NULL,
    amount numeric NOT NULL,
    created_at timestamp without time zone DEFAULT now() NOT NULL
);
CREATE SEQUENCE public.withdrawals_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
ALTER SEQUENCE public.withdrawals_id_seq OWNED BY public.withdrawals.id;
ALTER TABLE ONLY public.accounts ALTER COLUMN account_id SET DEFAULT nextval('public.accounts_account_id_seq'::regclass);
ALTER TABLE ONLY public.asset_accounts ALTER COLUMN id SET DEFAULT nextval('public.asset_accounts_id_seq'::regclass);
ALTER TABLE ONLY public.batch_history ALTER COLUMN id SET DEFAULT nextval('public.batch_history_id_seq'::regclass);
ALTER TABLE ONLY public.batch_history_details ALTER COLUMN id SET DEFAULT nextval('public.batch_history_details_id_seq'::regclass);
ALTER TABLE ONLY public.cached_metrics ALTER COLUMN id SET DEFAULT nextval('public.cached_metrics_id_seq'::regclass);
ALTER TABLE ONLY public.dropship_purchase_details ALTER COLUMN id SET DEFAULT nextval('public.dropship_purchase_details_id_seq'::regclass);
ALTER TABLE ONLY public.expense_lines ALTER COLUMN line_id SET DEFAULT nextval('public.expense_lines_line_id_seq'::regclass);
ALTER TABLE ONLY public.jenis_channels ALTER COLUMN jenis_channel_id SET DEFAULT nextval('public.jenis_channels_jenis_channel_id_seq'::regclass);
ALTER TABLE ONLY public.journal_entries ALTER COLUMN journal_id SET DEFAULT nextval('public.journal_entries_journal_id_seq'::regclass);
ALTER TABLE ONLY public.journal_lines ALTER COLUMN line_id SET DEFAULT nextval('public.journal_lines_line_id_seq'::regclass);
ALTER TABLE ONLY public.reconciled_transactions ALTER COLUMN id SET DEFAULT nextval('public.reconciled_transactions_id_seq'::regclass);
ALTER TABLE ONLY public.shopee_adjustments ALTER COLUMN id SET DEFAULT nextval('public.shopee_adjustments_id_seq'::regclass);
ALTER TABLE ONLY public.shopee_order_items ALTER COLUMN id SET DEFAULT nextval('public.shopee_order_items_id_seq'::regclass);
ALTER TABLE ONLY public.shopee_order_packages ALTER COLUMN id SET DEFAULT nextval('public.shopee_order_packages_id_seq'::regclass);
ALTER TABLE ONLY public.shopee_settled_orders ALTER COLUMN id SET DEFAULT nextval('public.shopee_settled_orders_id_seq'::regclass);
ALTER TABLE ONLY public.stores ALTER COLUMN store_id SET DEFAULT nextval('public.stores_store_id_seq'::regclass);
ALTER TABLE ONLY public.withdrawals ALTER COLUMN id SET DEFAULT nextval('public.withdrawals_id_seq'::regclass);
ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_account_code_key UNIQUE (account_code);
ALTER TABLE ONLY public.accounts
    ADD CONSTRAINT accounts_pkey PRIMARY KEY (account_id);
ALTER TABLE ONLY public.ad_invoices
    ADD CONSTRAINT ad_invoices_pkey PRIMARY KEY (invoice_no);
ALTER TABLE ONLY public.asset_accounts
    ADD CONSTRAINT asset_accounts_account_id_key UNIQUE (account_id);
ALTER TABLE ONLY public.asset_accounts
    ADD CONSTRAINT asset_accounts_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.batch_history_details
    ADD CONSTRAINT batch_history_details_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.batch_history
    ADD CONSTRAINT batch_history_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.cached_metrics
    ADD CONSTRAINT cached_metrics_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.cached_metrics
    ADD CONSTRAINT cached_metrics_shop_username_period_key UNIQUE (shop_username, period);
ALTER TABLE ONLY public.dropship_purchase_details
    ADD CONSTRAINT dropship_purchase_details_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.dropship_purchases
    ADD CONSTRAINT dropship_purchases_pkey PRIMARY KEY (kode_pesanan);
ALTER TABLE ONLY public.expense_lines
    ADD CONSTRAINT expense_lines_pkey PRIMARY KEY (line_id);
ALTER TABLE ONLY public.expenses
    ADD CONSTRAINT expenses_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.jenis_channels
    ADD CONSTRAINT jenis_channels_jenis_channel_key UNIQUE (jenis_channel);
ALTER TABLE ONLY public.jenis_channels
    ADD CONSTRAINT jenis_channels_pkey PRIMARY KEY (jenis_channel_id);
ALTER TABLE ONLY public.journal_entries
    ADD CONSTRAINT journal_entries_pkey PRIMARY KEY (journal_id);
ALTER TABLE ONLY public.journal_lines
    ADD CONSTRAINT journal_lines_pkey PRIMARY KEY (line_id);
ALTER TABLE ONLY public.reconciled_transactions
    ADD CONSTRAINT reconciled_transactions_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.shopee_adjustments
    ADD CONSTRAINT shopee_adjustments_no_pesanan_tanggal_penyesuaian_tipe_peny_key UNIQUE (no_pesanan, tanggal_penyesuaian, tipe_penyesuaian);
ALTER TABLE ONLY public.shopee_adjustments
    ADD CONSTRAINT shopee_adjustments_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.shopee_order_details
    ADD CONSTRAINT shopee_order_details_pkey PRIMARY KEY (order_sn);
ALTER TABLE ONLY public.shopee_order_items
    ADD CONSTRAINT shopee_order_items_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.shopee_order_packages
    ADD CONSTRAINT shopee_order_packages_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.shopee_settled_orders
    ADD CONSTRAINT shopee_settled_orders_order_id_key UNIQUE (order_id);
ALTER TABLE ONLY public.shopee_settled_orders
    ADD CONSTRAINT shopee_settled_orders_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.stores
    ADD CONSTRAINT stores_pkey PRIMARY KEY (store_id);
ALTER TABLE ONLY public.tax_payments
    ADD CONSTRAINT tax_payments_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.withdrawals
    ADD CONSTRAINT withdrawals_pkey PRIMARY KEY (id);
CREATE INDEX idx_batch_history_process_started ON public.batch_history USING btree (process_type, started_at DESC);
CREATE INDEX idx_dropship_purchase_details_kode ON public.dropship_purchase_details USING btree (kode_pesanan);
CREATE INDEX idx_dropship_purchases_composite ON public.dropship_purchases USING btree (jenis_channel, nama_toko, waktu_pesanan_terbuat DESC);
CREATE INDEX idx_dropship_purchases_invoice_channel ON public.dropship_purchases USING btree (kode_invoice_channel);
CREATE INDEX idx_expenses_date ON public.expenses USING btree (date DESC);
CREATE INDEX idx_journal_entries_created_at ON public.journal_entries USING btree (created_at DESC);
CREATE INDEX idx_journal_lines_account_id ON public.journal_lines USING btree (account_id);
CREATE INDEX idx_journal_lines_composite ON public.journal_lines USING btree (journal_id, account_id, amount, is_debit);
CREATE INDEX idx_purchases_pending_reconcile ON public.dropship_purchases USING btree (kode_invoice_channel, status_pesanan_terakhir) WHERE (status_pesanan_terakhir <> 'pesanan selesai'::text);
CREATE INDEX idx_shopee_settled_orders_order_id ON public.shopee_settled_orders USING btree (order_id);
CREATE UNIQUE INDEX journal_entries_source_idx ON public.journal_entries USING btree (source_type, source_id);
ALTER TABLE ONLY public.batch_history_details
    ADD CONSTRAINT batch_history_details_batch_id_fkey FOREIGN KEY (batch_id) REFERENCES public.batch_history(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.expense_lines
    ADD CONSTRAINT expense_lines_expense_id_fkey FOREIGN KEY (expense_id) REFERENCES public.expenses(id) ON DELETE CASCADE;
ALTER TABLE ONLY public.shopee_order_items
    ADD CONSTRAINT shopee_order_items_order_sn_fkey FOREIGN KEY (order_sn) REFERENCES public.shopee_order_details(order_sn) ON DELETE CASCADE;
ALTER TABLE ONLY public.shopee_order_packages
    ADD CONSTRAINT shopee_order_packages_order_sn_fkey FOREIGN KEY (order_sn) REFERENCES public.shopee_order_details(order_sn) ON DELETE CASCADE;
