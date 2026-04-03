# Updating the Map Data (CSV Instructions)

## Important — do this before exporting

Before you download the CSV, make sure the formula-based helper columns are filled all the way down for every active row.

Do not forget to drag down these columns if needed:

- Customer System Total
- Matched Customer Name
- Matched Customer Type
- Matched Latitude
- Matched Longitude
- Matched Population

If these formulas are not filled down correctly, the map may load incomplete or incorrect names, types, coordinates, or population values.

---

The map reads its data from a CSV file exported from our master spreadsheet.
To update the map, follow these steps exactly.

---

## Step 1 — Open the spreadsheet

Open the master spreadsheet here:

https://docs.google.com/spreadsheets/d/1ASYXQ3Bdt0FWHPqG7nQ1ZVKDCzk0Lk2CsHiTHHP4jaA/edit?gid=0#gid=0

---

## Step 2 — Export the sheet as a CSV

In Google Sheets:

1. Click **File**
2. Click **Download**
3. Click **Comma Separated Values (.csv)**

Google will download the file to your computer.

---

## Step 3 — Rename the file (IMPORTANT)

The filename must exactly match:

**Impact_Map_Export - System Bridge (Anchor point).csv**

This includes:

- spaces
- capitalization
- punctuation
- parentheses
- the dash
- the `.csv` extension

If the filename does **not** match exactly, the map will not load the data.

Examples of incorrect filenames:

- Impact_Map_Export.csv
- Impact Map Export System Bridge.csv
- Impact_Map_Export - System Bridge (Anchor point) (1).csv
- impact_map_export - system bridge (anchor point).csv

These will break the map.

---

## Step 4 — Replace the old CSV file

Put the renamed file into the same folder as:

**index.html**

Replace the existing CSV file when prompted.

---

## Step 5 — Upload changes to GitHub (if using GitHub Pages)

If the site is hosted on GitHub Pages:

1. Upload the new CSV file
2. Commit the change
3. Wait about **30 seconds to 5 minutes** for the map to update
4. Refresh the map page

If the changes do not appear right away, your browser may still be showing an older cached version. Try refreshing the page again or opening the map in a different browser.

The updates should now appear on the map.

---

Created by Sean Ryan for All We Are
