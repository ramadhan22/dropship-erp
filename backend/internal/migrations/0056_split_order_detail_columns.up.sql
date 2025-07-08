ALTER TABLE shopee_order_details
    ADD COLUMN status TEXT,
    ADD COLUMN checkout_time BIGINT,
    ADD COLUMN update_time BIGINT,
    ADD COLUMN pay_time BIGINT,
    ADD COLUMN total_amount NUMERIC,
    ADD COLUMN currency TEXT;

ALTER TABLE shopee_order_items
    ADD COLUMN order_item_id BIGINT,
    ADD COLUMN item_name TEXT,
    ADD COLUMN model_original_price NUMERIC,
    ADD COLUMN model_quantity_purchased INT;

UPDATE shopee_order_details SET
    status = detail->>'status',
    checkout_time = (detail->>'checkout_time')::bigint,
    update_time = (detail->>'update_time')::bigint,
    pay_time = (detail->>'pay_time')::bigint,
    total_amount = (detail->>'total_amount')::numeric,
    currency = detail->>'currency';

UPDATE shopee_order_items SET
    order_item_id = (item->>'order_item_id')::bigint,
    item_name = item->>'item_name',
    model_original_price = (item->>'model_original_price')::numeric,
    model_quantity_purchased = (item->>'model_quantity_purchased')::int;

ALTER TABLE shopee_order_details DROP COLUMN detail;
ALTER TABLE shopee_order_items DROP COLUMN item;
