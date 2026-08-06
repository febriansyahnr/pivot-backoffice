-- +goose Up
-- +goose StatementBegin
CREATE TABLE industries (
    uuid CHAR(36) PRIMARY KEY,
    parent_industry VARCHAR(255) NOT NULL,
    child_industry VARCHAR(255) NOT NULL,
    risk_level VARCHAR(50) NOT NULL,
    mcc VARCHAR(10) NOT NULL,
    common_mcc VARCHAR(255) NOT NULL,
    created_at TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6),
    updated_at TIMESTAMP(6) DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6)
);

-- Insert industry data
INSERT INTO industries (uuid, parent_industry, child_industry, risk_level, mcc, common_mcc) VALUES
-- Airlines
(UUID(), 'Airlines', 'Airlines, Air Carriers', 'High', '4511', '4511'),

-- Automotive
(UUID(), 'Automotive', 'Car and Truck Dealers', 'Medium', '5511', '7699'),
(UUID(), 'Automotive', 'Automotive Parts and Accessories Stores', 'Medium', '5533', '7699'),
(UUID(), 'Automotive', 'Service Stations (with or without Ancillary Services)', 'Medium', '5541', '7699'),

-- Clothing, apparel & accessories
(UUID(), 'Clothing, apparel & acessories', 'Clothing stores', 'Low', '5651', '7299'),
(UUID(), 'Clothing, apparel & acessories', 'Pet shop', 'Low', '5995', '7299'),
(UUID(), 'Clothing, apparel & acessories', 'Tailors', 'Low', '5697', '7299'),
(UUID(), 'Clothing, apparel & acessories', 'Men''s and Women''s Clothing Stores', 'Low', '5691', '7299'),

-- Digital goods
(UUID(), 'Digital goods', 'Games', 'Medium', '5816', '5816'),
(UUID(), 'Digital goods', 'Media (books, movies, artwork, images)', 'Medium', '5815', '5815'),
(UUID(), 'Digital goods', 'Applications (excl. games)', 'Medium', '5817', '5817'),
(UUID(), 'Digital goods', 'Utilities', 'Medium', '4900', '4900'),

-- Education
(UUID(), 'Education', 'Elementary and Secondary Schools', 'Low', '8211', '8299'),
(UUID(), 'Education', 'Colleges, Universities, Professional Schools', 'Low', '8220', '8299'),
(UUID(), 'Education', 'Correspondence Schools', 'Low', '8241', '8299'),
(UUID(), 'Education', 'Other educational services', 'Low', '8299', '8299'),
(UUID(), 'Education', 'LMS Platform', 'Low', '8299', '8299'),
(UUID(), 'Education', 'Edutech', 'Low', '8299', '8299'),

-- Entertainment
(UUID(), 'Entertainment', 'Theaters', 'Low', '7832', '7999'),
(UUID(), 'Entertainment', 'Tourist Attractions and Exhibits', 'Medium', '7991', '7999'),
(UUID(), 'Entertainment', 'Public Golf Courses', 'Low', '7992', '7999'),
(UUID(), 'Entertainment', 'Membership Clubs (Sports, Recreation)', 'Medium', '7997', '7999'),
(UUID(), 'Entertainment', 'Streaming services (gaming, music, TV)', 'Medium', '4899', '7999'),
(UUID(), 'Entertainment', 'Health & beauty spas', 'Medium', '7298', '7999'),
(UUID(), 'Entertainment', 'Other entertainment/recreation services', 'Medium', '7999', '7999'),

-- Financial services
(UUID(), 'Financial services', 'Banks, Credit unions', 'High', '6010', '6010'),
(UUID(), 'Financial services', 'Remittance', 'High', '4829', '4829'),
(UUID(), 'Financial services', 'Quasi-Cash Transactions (Gambling, Lottery)', 'High', '6051', '6051'),
(UUID(), 'Financial services', 'Investments', 'High', '6211', '6211'),
(UUID(), 'Financial services', 'Cryptocurrency exchage', 'High', '6051', '6051'),
(UUID(), 'Financial services', 'Forex', 'High', '6051', '6051'),
(UUID(), 'Financial services', 'P2P Lending', 'High', '6051', '6051'),
(UUID(), 'Financial services', 'Other financial services', 'High', '6051', '6051'),
(UUID(), 'Financial services', 'Aggregator/Payment Reseller', 'High', '6051', '6051'),

-- Healthcare
(UUID(), 'Healthcare', 'Hospital', 'Medium', '8062', '8099'),
(UUID(), 'Healthcare', 'Clinics (Chiropractors, Dentist, Optometrics)', 'Medium', '8099', '8099'),
(UUID(), 'Healthcare', 'Drug stores/Pharmacy', 'Medium', '5912', '8099'),
(UUID(), 'Healthcare', 'Laboratories', 'Medium', '8099', '8099'),
(UUID(), 'Healthcare', 'Opticians, Optical Goods, Eyeglasses', 'Low', '8043', '8099'),

-- Logistics
(UUID(), 'Logistics', 'Courier, Express, and Parcel', 'Medium', '4215', '4789'),
(UUID(), 'Logistics', 'Third-Party Logistics (3PL) Provider', 'Medium', '4215', '4789'),
(UUID(), 'Logistics', 'Freight Forwarding Companies & Multimodal Transport Operators (MTOs)', 'Medium', '4215', '4789'),
(UUID(), 'Logistics', 'Cold Chain Logistics Providers', 'Medium', '4225', '4789'),

-- Marketplace
(UUID(), 'Marketplace', 'Horizontal Marketplace', 'Medium', '5262', '5262'),
(UUID(), 'Marketplace', 'Vertical Marketplace', 'Medium', '5262', '5262'),
(UUID(), 'Marketplace', 'Gaming Marketplace', 'High', '5262', '5262'),

-- Organization
(UUID(), 'Organization', 'Charitable and Social Service Organizations', 'High', '8398', '8699'),
(UUID(), 'Organization', 'Political organizations', 'High', '8651', '8699'),
(UUID(), 'Organization', 'Religious organizations', 'Low', '8661', '8699'),
(UUID(), 'Organization', 'Civic, Social, and Fraternal Associations', 'Low', '8641', '8699'),

-- Outsourcing
(UUID(), 'Outsourcing', 'Freelance marketplace/Gig economy platform/Crowdsourcing', 'Medium', '7999', '7399'),
(UUID(), 'Outsourcing', 'Business Process Outsourcing (BPO)', 'Medium', '7999', '7399'),

-- Personal services
(UUID(), 'Personal services', 'Funeral services/crematorium', 'Low', '7261', '7399'),
(UUID(), 'Personal services', 'Beauty & barber shops', 'Low', '7230', '7399'),
(UUID(), 'Personal services', 'Laundry, cleaning, garment services', 'Low', '7210', '7399'),
(UUID(), 'Personal services', 'Photography Studios', 'Low', '7221', '7399'),
(UUID(), 'Personal services', 'Wedding and Bridal Services', 'Low', '7230', '7399'),
(UUID(), 'Personal services', 'Counseling services', 'Low', '7277', '7399'),
(UUID(), 'Personal services', 'Massage parlors', 'Low', '7297', '7399'),

-- Professional services
(UUID(), 'Professional services', 'Advertising Services', 'Low', '7311', '8999'),
(UUID(), 'Professional services', 'Commercial Photography, Art, and Graphics', 'Low', '7333', '8999'),
(UUID(), 'Professional services', 'Consulting, Public Relations Services', 'Low', '7392', '8999'),
(UUID(), 'Professional services', 'Professional Services (Not Elsewhere Classified)', 'Low', '8999', '8999'),
(UUID(), 'Professional services', 'Law firm', 'Medium', '8111', '8999'),
(UUID(), 'Professional services', 'Accounting, Auditing, Book keeping', 'Medium', '8931', '8999'),
(UUID(), 'Professional services', 'Insurance Sales, Underwriting, and Premiums', 'High', '6300', '8999'),
(UUID(), 'Professional services', 'Timeshares', 'Medium', '7012', '8999'),
(UUID(), 'Professional services', 'Tax Preparation Services', 'Low', '7276', '8999'),
(UUID(), 'Professional services', 'Counseling Services – Debt, Marriage, and Personal', 'Low', '7277', '8999'),
(UUID(), 'Professional services', 'Employment Agencies and Temporary Help Services', 'Low', '7361', '8999'),
(UUID(), 'Professional services', 'Management, Consulting, and Public Relations Services', 'Medium', '7392', '8999'),
(UUID(), 'Professional services', 'Detective Agencies, Protective Services, and Security Services, including Armored Cars, and Guard Dogs', 'Low', '7393', '8999'),
(UUID(), 'Professional services', 'Architectural, Engineering, and Surveying Services', 'Low', '8911', '8999'),

-- Recreational services
(UUID(), 'Recreational services', 'Fitness & Sports Club', 'Medium', '7997', '8999'),

-- Restaurants
(UUID(), 'Restaurants', 'Restaurants and Eating Places', 'Low', '5812', '5812'),
(UUID(), 'Restaurants', 'Drinking Places', 'Low', '5813', '5812'),
(UUID(), 'Restaurants', 'Coffee shops/cafe', 'Low', '5812', '5812'),
(UUID(), 'Restaurants', 'Fast Food Restaurants', 'Low', '5814', '5812'),

-- Retail
(UUID(), 'Retail', 'Department Stores', 'Low', '5311', '5999'),
(UUID(), 'Retail', 'Grocery Stores', 'Low', '5411', '5999'),
(UUID(), 'Retail', 'Miscellaneous and Specialty Retail', 'Low', '5999', '5999'),
(UUID(), 'Retail', 'Book Stores', 'Low', '5942', '5999'),
(UUID(), 'Retail', 'Office Supplies', 'Low', '5943', '5999'),
(UUID(), 'Retail', 'Furniture, home decor & home appliances', 'Low', '5712', '5999'),
(UUID(), 'Retail', 'Alcohol', 'High', '5921', '5999'),
(UUID(), 'Retail', 'Other retail', 'Medium', '5999', '5999'),

-- SaaS
(UUID(), 'SaaS', 'POS', 'Low', '7399', '7399'),
(UUID(), 'SaaS', 'CRM & Marketing Automation', 'Low', '7399', '7399'),
(UUID(), 'SaaS', 'HRIS', 'Low', '7399', '7399'),
(UUID(), 'SaaS', 'Invoicing platform', 'Low', '7399', '7399'),
(UUID(), 'SaaS', 'Enabler (website developer, ecommerce enabler)', 'Medium', '4816', '4816'),
(UUID(), 'SaaS', 'Ecommerce enabler', 'Medium', '4816', '4816'),

-- Travel services
(UUID(), 'Travel services', 'Lodging - Hotels, Motels, Resorts', 'Medium', '7011', '7011'),
(UUID(), 'Travel services', 'Online travel agent', 'Medium', '4722', '7011');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS industries;
-- +goose StatementEnd
