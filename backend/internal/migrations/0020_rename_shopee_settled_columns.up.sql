-- 0020_rename_shopee_settled_columns.up.sql
-- Rename columns in shopee_settled to match new headers
ALTER TABLE shopee_settled RENAME COLUMN komisi_shopee TO diskon_produk_dari_shopee;
ALTER TABLE shopee_settled RENAME COLUMN biaya_admin_shopee TO diskon_voucher_ditanggung_penjual;
ALTER TABLE shopee_settled RENAME COLUMN biaya_layanan TO cashback_koin_ditanggung_penjual;
ALTER TABLE shopee_settled RENAME COLUMN biaya_layanan_ekstra TO ongkir_dibayar_pembeli;
ALTER TABLE shopee_settled RENAME COLUMN biaya_penyedia_pembayaran TO diskon_ongkir_ditanggung_jasa_kirim;
ALTER TABLE shopee_settled RENAME COLUMN asuransi TO gratis_ongkir_dari_shopee;
ALTER TABLE shopee_settled RENAME COLUMN total_biaya_transaksi TO ongkir_yang_diteruskan_oleh_shopee_ke_jasa_kirim;
ALTER TABLE shopee_settled RENAME COLUMN biaya_pengiriman TO ongkos_kirim_pengembalian_barang;
ALTER TABLE shopee_settled RENAME COLUMN total_diskon_pengiriman TO pengembalian_biaya_kirim;
ALTER TABLE shopee_settled RENAME COLUMN promo_gratis_ongkir_shopee TO biaya_komisi_ams;
ALTER TABLE shopee_settled RENAME COLUMN promo_gratis_ongkir_penjual TO biaya_administrasi;
ALTER TABLE shopee_settled RENAME COLUMN promo_diskon_shopee TO biaya_layanan_termasuk_ppn_11;
ALTER TABLE shopee_settled RENAME COLUMN promo_diskon_penjual TO premi;
ALTER TABLE shopee_settled RENAME COLUMN cashback_shopee TO biaya_program;
ALTER TABLE shopee_settled RENAME COLUMN cashback_penjual TO biaya_kartu_kredit;
ALTER TABLE shopee_settled RENAME COLUMN koin_shopee TO biaya_kampanye;
ALTER TABLE shopee_settled RENAME COLUMN potongan_lainnya TO bea_masuk_ppn_pph;
ALTER TABLE shopee_settled RENAME COLUMN total_penerimaan TO total_penghasilan;
