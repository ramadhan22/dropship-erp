-- Insert 2025 expense data and related journal entries
DO $$
DECLARE
    kas INT := (SELECT account_id FROM accounts WHERE account_code='1.1.14');
    mitra INT := (SELECT account_id FROM accounts WHERE account_code='5.2.7');
    voucher_acc INT := (SELECT account_id FROM accounts WHERE account_code='5.2.3');
    iklan INT := (SELECT account_id FROM accounts WHERE account_code='5.2.9');
    layanan INT := (SELECT account_id FROM accounts WHERE account_code='5.2.4');
    gratis INT := (SELECT account_id FROM accounts WHERE account_code='5.2.6');
    cogs INT := (SELECT account_id FROM accounts WHERE account_code='5.1');
    loan INT := (SELECT account_id FROM accounts WHERE account_code='5.2.1');
    capex INT := (SELECT account_id FROM accounts WHERE account_code='5.2.8');
    ids UUID[] := ARRAY[
      '389a5e77-985c-4cb1-8be2-b16c254080b3'::uuid,
      '7e06266c-a895-4b42-9049-defc62176e6d'::uuid,
      '245eccdd-dc5c-4083-ab37-3087fc6c3851'::uuid,
      '447dcdd7-bf3f-46cd-a709-f649cbd6efa0'::uuid,
      '7e817fbb-c39e-4be6-b5ac-79774db864bc'::uuid,
      'afbf2275-555d-4673-b99d-6a59680dde17'::uuid,
      '2ac23cd1-e62c-4bdb-935e-4365a702deff'::uuid,
      '0187d429-3f4b-4b5a-9675-f580ad1462a7'::uuid,
      '0fc33c2c-8e08-4fb9-baf5-fdb97460ce82'::uuid,
      '91ef8400-84e8-433b-98dc-0e89cffcebf3'::uuid,
      'c5fdd8e7-d6b0-48c9-a846-5bc63ae3059f'::uuid,
      '9d330096-7e4c-4160-a8b1-1f84b3faa23c'::uuid,
      '8b83b93c-e92f-416b-990e-e0cda3c6e906'::uuid,
      '56607835-25c9-49cf-9f64-791ae0735187'::uuid,
      '9bafd483-c30a-42a7-a679-2cf6af172eb5'::uuid,
      '2d9cb21e-3cd6-45d2-8f38-31fd65dfd068'::uuid
    ];
    dates DATE[] := ARRAY[
      '2025-01-19','2025-01-20','2025-01-31','2025-01-31','2025-01-31',
      '2025-02-20','2025-03-06','2025-03-16','2025-03-28','2025-03-31',
      '2025-04-21','2025-04-30','2025-05-08','2025-05-13','2025-05-23','2025-05-24'
    ];
    notes TEXT[] := ARRAY[
      'Member','Diskon','Iklan Tokopedia','Biaya Layanan Tokopedia','Gratis Ongkir Tokopedia',
      'Member','Member','Iklan Tiktok','Barang Hilang','Dana Cepat',
      'Member','Dana Cepat','Member','Iklan Tiktok','Bikin Logo MR eStore','Kursi Kerja'
    ];
    amounts NUMERIC[] := ARRAY[
      388500,20100,165000,37215,23654,
      388500,375000,111000,-91700,51016,
      388500,221073,595700,300000,300000,1005280
    ];
    accts INT[] := ARRAY[
      mitra,voucher_acc,iklan,layanan,gratis,
      mitra,mitra,iklan,cogs,loan,
      mitra,loan,mitra,iklan,capex,capex
    ];
    i INT;
    jid INT;
BEGIN
  FOR i IN 1 .. array_length(ids,1) LOOP
    INSERT INTO expenses (id, date, description, amount, asset_account_id)
    VALUES (ids[i], dates[i], notes[i], amounts[i], kas);
    INSERT INTO expense_lines (expense_id, account_id, amount)
    VALUES (ids[i], accts[i], amounts[i]);
    INSERT INTO journal_entries (entry_date, description, source_type, source_id, shop_username, store, created_at)
    VALUES (dates[i], notes[i], 'expense', ids[i], '', '', now())
    RETURNING journal_id INTO jid;
    INSERT INTO journal_lines (journal_id, account_id, is_debit, amount, memo)
    VALUES (jid, accts[i], true, amounts[i], notes[i]);
    INSERT INTO journal_lines (journal_id, account_id, is_debit, amount, memo)
    VALUES (jid, kas, false, amounts[i], notes[i]);
  END LOOP;
END $$;
