# All We Are Impact Map

## How It Works

The map loads its data **directly from the published Google Sheet** every time someone visits the page. There is no need to manually export, rename, or upload CSV files.

The master spreadsheet is here:

https://docs.google.com/spreadsheets/d/1ASYXQ3Bdt0FWHPqG7nQ1ZVKDCzk0Lk2CsHiTHHP4jaA/edit?gid=0#gid=0

When you edit the spreadsheet, the map will reflect the changes automatically (Google caches the published data for roughly 5 minutes, so updates may not appear instantly).

---

## Important — Keep Helper Columns Filled

Before adding new rows, make sure the formula-based helper columns are dragged down for every active row:

- Customer System Total
- Matched Customer Name
- Matched Customer Type
- Matched Latitude
- Matched Longitude
- Matched Population

If these formulas are not filled down correctly, the map may show incomplete or incorrect data.

---

## Keeping the Sheet Published

The map relies on the Google Sheet being **published to the web** (File → Share → Publish to the web, CSV format). If publishing is ever turned off, the map will stop loading data. The sheet can still be restricted for editing — publishing only provides read-only access.

---

Created by Sean Ryan for All We Are
