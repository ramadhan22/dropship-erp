ALTER TABLE shopee_order_details ADD COLUMN detail JSONB;
ALTER TABLE shopee_order_items ADD COLUMN item JSONB;

UPDATE shopee_order_details SET
    detail = jsonb_build_object(
        'status', status,
        'checkout_time', checkout_time,
        'update_time', update_time,
        'pay_time', pay_time,
        'total_amount', total_amount,
        'currency', currency
    );

UPDATE shopee_order_items SET
    item = jsonb_build_object(
        'order_item_id', order_item_id,
        'item_name', item_name,
        'model_original_price', model_original_price,
        'model_quantity_purchased', model_quantity_purchased
    );

ALTER TABLE shopee_order_details
    DROP COLUMN status,
    DROP COLUMN checkout_time,
    DROP COLUMN update_time,
    DROP COLUMN pay_time,
    DROP COLUMN total_amount,
    DROP COLUMN currency;

ALTER TABLE shopee_order_items
    DROP COLUMN order_item_id,
    DROP COLUMN item_name,
    DROP COLUMN model_original_price,
    DROP COLUMN model_quantity_purchased;
