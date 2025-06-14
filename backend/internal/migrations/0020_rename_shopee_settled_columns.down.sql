-- 0020_rename_shopee_settled_columns.down.sql
-- Revert column renames in shopee_settled
ALTER TABLE shopee_settled RENAME COLUMN diskon_produk_dari_shopee TO komisi_shopee;
ALTER TABLE shopee_settled RENAME COLUMN diskon_voucher_ditanggung_penjual TO biaya_admin_shopee;
ALTER TABLE shopee_settled RENAME COLUMN cashback_koin_ditanggung_penjual TO biaya_layanan;
ALTER TABLE shopee_settled RENAME COLUMN ongkir_dibayar_pembeli TO biaya_layanan_ekstra;
ALTER TABLE shopee_settled RENAME COLUMN diskon_ongkir_ditanggung_jasa_kirim TO biaya_penyedia_pembayaran;
ALTER TABLE shopee_settled RENAME COLUMN gratis_ongkir_dari_shopee TO asuransi;
ALTER TABLE shopee_settled RENAME COLUMN ongkir_yang_diteruskan_oleh_shopee_ke_jasa_kirim TO total_biaya_transaksi;
ALTER TABLE shopee_settled RENAME COLUMN ongkos_kirim_pengembalian_barang TO biaya_pengiriman;
ALTER TABLE shopee_settled RENAME COLUMN pengembalian_biaya_kirim TO total_diskon_pengiriman;
ALTER TABLE shopee_settled RENAME COLUMN biaya_komisi_ams TO promo_gratis_ongkir_shopee;
ALTER TABLE shopee_settled RENAME COLUMN biaya_administrasi TO promo_gratis_ongkir_penjual;
ALTER TABLE shopee_settled RENAME COLUMN biaya_layanan_termasuk_ppn_11 TO promo_diskon_shopee;
ALTER TABLE shopee_settled RENAME COLUMN premi TO promo_diskon_penjual;
ALTER TABLE shopee_settled RENAME COLUMN biaya_program TO cashback_shopee;
ALTER TABLE shopee_settled RENAME COLUMN biaya_kartu_kredit TO cashback_penjual;
ALTER TABLE shopee_settled RENAME COLUMN biaya_kampanye TO koin_shopee;
ALTER TABLE shopee_settled RENAME COLUMN bea_masuk_ppn_pph TO potongan_lainnya;
ALTER TABLE shopee_settled RENAME COLUMN total_penghasilan TO total_penerimaan;
