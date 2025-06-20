-- United States
INSERT INTO tax_rates (country, state, inclusive, percentage)
VALUES
('US', 'AL', FALSE, 921),  -- Alabama
('US', 'AK', FALSE, 176),  -- Alaska (no state tax, local only)
('US', 'AZ', FALSE, 838),  -- Arizona
('US', 'AR', FALSE, 944),  -- Arkansas
('US', 'CA', FALSE, 882),  -- California
('US', 'CO', FALSE, 777),  -- Colorado
('US', 'CT', FALSE, 635),  -- Connecticut
('US', 'DE', FALSE, 0),    -- Delaware
('US', 'FL', FALSE, 702),  -- Florida
('US', 'GA', FALSE, 737),  -- Georgia
('US', 'HI', FALSE, 444),  -- Hawaii (GET, not traditional sales tax)
('US', 'ID', FALSE, 606),  -- Idaho
('US', 'IL', FALSE, 882),  -- Illinois
('US', 'IN', FALSE, 700),  -- Indiana
('US', 'IA', FALSE, 688),  -- Iowa
('US', 'KS', FALSE, 869),  -- Kansas
('US', 'KY', FALSE, 600),  -- Kentucky
('US', 'LA', FALSE, 988),  -- Louisiana
('US', 'ME', FALSE, 555),  -- Maine
('US', 'MD', FALSE, 600),  -- Maryland
('US', 'MA', FALSE, 625),  -- Massachusetts
('US', 'MI', FALSE, 600),  -- Michigan
('US', 'MN', FALSE, 741),  -- Minnesota
('US', 'MS', FALSE, 707),  -- Mississippi
('US', 'MO', FALSE, 820),  -- Missouri
('US', 'MT', FALSE, 0),    -- Montana
('US', 'NE', FALSE, 691),  -- Nebraska
('US', 'NV', FALSE, 836),  -- Nevada
('US', 'NH', FALSE, 0),    -- New Hampshire
('US', 'NJ', FALSE, 666),  -- New Jersey
('US', 'NM', FALSE, 785),  -- New Mexico
('US', 'NY', FALSE, 852),  -- New York
('US', 'NC', FALSE, 698),  -- North Carolina
('US', 'ND', FALSE, 698),  -- North Dakota
('US', 'OH', FALSE, 724),  -- Ohio
('US', 'OK', FALSE, 947),  -- Oklahoma
('US', 'OR', FALSE, 0),    -- Oregon (no sales tax)
('US', 'PA', FALSE, 634),  -- Pennsylvania
('US', 'RI', FALSE, 700),  -- Rhode Island
('US', 'SC', FALSE, 747),  -- South Carolina
('US', 'SD', FALSE, 642),  -- South Dakota
('US', 'TN', FALSE, 955),  -- Tennessee
('US', 'TX', FALSE, 820),  -- Texas
('US', 'UT', FALSE, 712),  -- Utah
('US', 'VT', FALSE, 620),  -- Vermont
('US', 'VA', FALSE, 581),  -- Virginia
('US', 'WA', FALSE, 940),  -- Washington
('US', 'WV', FALSE, 654),  -- West Virginia
('US', 'WI', FALSE, 543),  -- Wisconsin
('US', 'WY', FALSE, 539),  -- Wyoming
('US', 'DC', FALSE, 620),  -- District of Columbia

-- U.S. Territories
('US', 'AS', FALSE, 0),    -- American Samoa
('US', 'GU', FALSE, 400),  -- Guam
('US', 'MP', FALSE, 0),    -- Northern Mariana Islands
('US', 'PR', FALSE, 1050), -- Puerto Rico
('US', 'VI', FALSE, 600);  -- U.S. Virgin Islands

INSERT INTO tax_rates (country, state, inclusive, percentage)
VALUES
('CA', 'AB', TRUE, 500),   -- Alberta
('CA', 'BC', TRUE, 1200),  -- British Columbia
('CA', 'MB', TRUE, 1200),  -- Manitoba
('CA', 'NB', TRUE, 1500),  -- New Brunswick
('CA', 'NL', TRUE, 1500),  -- Newfoundland and Labrador
('CA', 'NS', TRUE, 1500),  -- Nova Scotia
('CA', 'NT', TRUE, 500),   -- Northwest Territories
('CA', 'NU', TRUE, 500),   -- Nunavut
('CA', 'ON', TRUE, 1300),  -- Ontario
('CA', 'PE', TRUE, 1500),  -- Prince Edward Island
('CA', 'QC', TRUE, 1498),  -- Quebec
('CA', 'SK', TRUE, 1100),  -- Saskatchewan
('CA', 'YT', TRUE, 500);   -- Yukon (GST only)

-- Scandinavia
INSERT INTO tax_rates (country, inclusive, percentage)
VALUES
('SE', TRUE, 2500), -- Sweden
('NO', TRUE, 2500), -- Norway
('DK', TRUE, 2500), -- Denmark
('IS', TRUE, 2400), -- Iceland
('FI', TRUE, 2400); -- Finland

INSERT INTO tax_rates (country, inclusive, percentage)
VALUES
-- Western Europe
('GB', TRUE, 2000),  -- United Kingdom
('CH', TRUE, 770),   -- Switzerland
('LI', TRUE, 770),   -- Liechtenstein
('DE', TRUE, 1900),  -- Germany
('FR', TRUE, 2000),  -- France
('IT', TRUE, 2200),  -- Italy
('ES', TRUE, 2100),  -- Spain
('NL', TRUE, 2100),  -- Netherlands
('BE', TRUE, 2100),  -- Belgium
('AT', TRUE, 2000),  -- Austria
('LU', TRUE, 1600),  -- Luxembourg
('IE', TRUE, 2300),  -- Ireland
('PT', TRUE, 2300),  -- Portugal

-- Central Europe
('PL', TRUE, 2300),  -- Poland
('CZ', TRUE, 2100),  -- Czech Republic
('SK', TRUE, 2000),  -- Slovakia
('HU', TRUE, 2700),  -- Hungary (highest in EU)
('SI', TRUE, 2200),  -- Slovenia
('HR', TRUE, 2500),  -- Croatia

-- Eastern Europe & Balkans
('RO', TRUE, 1900),  -- Romania
('BG', TRUE, 2000),  -- Bulgaria
('EE', TRUE, 2000),  -- Estonia
('LV', TRUE, 2100),  -- Latvia
('LT', TRUE, 2100),  -- Lithuania

-- Southeast & neighboring countries
('RS', TRUE, 2000),  -- Serbia
('BA', TRUE, 1700),  -- Bosnia & Herzegovina
('MK', TRUE, 1800),  -- North Macedonia
('MD', TRUE, 2000),  -- Moldova
('UA', TRUE, 2000),  -- Ukraine

-- Asia Pacific
('JP', TRUE, 1000),  -- Japan
('SG', TRUE, 900),   -- Singapore
('KR', TRUE, 1000),  -- South Korea
('IN', TRUE, 1800),  -- India: ~18% GST average
('CN', TRUE, 1300),  -- China

-- Middle East
('AE', TRUE, 500),   -- UAE
('SA', TRUE, 1500),  -- Saudi Arabia
('IL', TRUE, 1700),  -- Israel

-- Africa
('ZA', TRUE, 1500),  -- South Africa
('NG', TRUE, 750),   -- Nigeria

-- Latin America
('BR', TRUE, 1700),  -- Brazil
('MX', TRUE, 1600),  -- Mexico
('AR', TRUE, 2100),  -- Argentina
('CL', TRUE, 1900),  -- Chile

-- Oceania
('AU', TRUE, 1000),  -- Australia
('NZ', TRUE, 1500);  -- New Zealand
