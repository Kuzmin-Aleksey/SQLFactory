CREATE TABLE orders
(
    city_id                      INT,
    order_id                     VARCHAR(16),
    tender_id                    VARCHAR(16),
    user_id                      VARCHAR(16),
    driver_id                    VARCHAR(16),
    offset_hours                 INT,
    status_order                 VARCHAR(16),
    status_tender                VARCHAR(16),
    order_timestamp              TIMESTAMP,
    tender_timestamp             TIMESTAMP,
    driveraccept_timestamp       TIMESTAMP,
    driverarrived_timestamp      TIMESTAMP,
    driverstarttheride_timestamp TIMESTAMP,
    driverdone_timestamp         TIMESTAMP,
    clientcancel_timestamp       TIMESTAMP,
    drivercancel_timestamp       TIMESTAMP,
    order_modified_local         TIMESTAMP,
    cancel_before_accept_local   TIMESTAMP,
    distance_in_meters           INT,
    duration_in_seconds          INT,
    price_order_local            DECIMAL(15, 2),
    price_tender_local           DECIMAL(15, 2),
    price_start_local            DECIMAL(15, 2)

);

-- Копирование данных из csv в таблицу
COPY orders (city_id, order_id, tender_id, user_id, driver_id, offset_hours, status_order, status_tender, order_timestamp,
           tender_timestamp, driveraccept_timestamp, driverarrived_timestamp, driverstarttheride_timestamp,
           driverdone_timestamp, clientcancel_timestamp, drivercancel_timestamp, order_modified_local,
           cancel_before_accept_local, distance_in_meters, duration_in_seconds, price_order_local, price_tender_local,
           price_start_local)
    FROM '/testdata/orders.csv'
    DELIMITER ','
    CSV HEADER;