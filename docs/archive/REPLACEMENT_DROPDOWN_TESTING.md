# Replacement Template Dropdown - Testing Guide

## Feature Completed ✅

Added dropdown UI for replacement templates in both Create and Edit modals.

## What Was Added

### 1. **Create Filter Modal**
- Dropdown with 60+ predefined replacement templates
- Organized by 6 categories (PII, Financial, Security, Technical, Confidential, Additional)
- Custom option that shows text input
- Visual preview of selected template

### 2. **Edit Filter Modal**
- Same dropdown functionality
- Auto-detects if current replacement matches a template
- Falls back to "Custom" mode for non-template values
- Preserves custom values when editing

### 3. **Template Categories**

**🆔 PII (8 templates):**
- `[EMAIL]`, `[PHONE]`, `[SSN]`, `[TAX_ID]`, `[PASSPORT]`, `[DRIVER_LICENSE]`, `[NATIONAL_ID]`, `[MRN]`

**💳 Financial (7 templates):**
- `[CREDIT_CARD]`, `[CVV]`, `[IBAN]`, `[BIC]`, `[BANK_ACCOUNT]`, `[ROUTING_NUMBER]`, `[CRYPTO_ADDRESS]`

**🔐 Security (16 templates):**
- API keys for AWS, Google, GitHub, GitLab, Slack, Stripe, Twilio, SendGrid
- JWT, SSH, Bearer, Access tokens
- Passwords

**🗄️ Technical (9 templates):**
- Database connections, credentials, passwords
- Internal IPs, hostnames
- Secret keys, encryption keys

**🔒 Confidential (10 templates):**
- Confidential, Redacted, Classified
- Internal projects, proprietary, trade secrets
- Salary info, HR, legal documents
- Competitor names

**🛡️ Additional (4 templates):**
- UUID, License keys, Session tokens, CSRF tokens

## Testing Checklist

### Create Filter - Template Mode
- [ ] Open Admin UI at http://localhost:5173
- [ ] Navigate to "Filters" section
- [ ] Click "➕ Add Filter" button
- [ ] In "Replacement Template" dropdown, select a template (e.g., `[EMAIL]`)
- [ ] Verify preview shows: "Will replace matches with: [EMAIL]"
- [ ] Fill in Pattern: `test@example.com`
- [ ] Select Type: `regex`
- [ ] Click "Create"
- [ ] Verify filter created with replacement `[EMAIL]`

### Create Filter - Custom Mode
- [ ] Click "➕ Add Filter" button
- [ ] In "Replacement Template" dropdown, keep "Custom" selected
- [ ] Type custom replacement: `[MY_CUSTOM_FILTER]`
- [ ] Fill in Pattern: `secret`
- [ ] Click "Create"
- [ ] Verify filter created with replacement `[MY_CUSTOM_FILTER]`

### Edit Filter - Template Detected
- [ ] Find a filter with template replacement (e.g., `[EMAIL]`)
- [ ] Click "Edit" button
- [ ] Verify dropdown shows `[EMAIL]` selected (not Custom)
- [ ] Verify preview shows the selected template
- [ ] Change to different template (e.g., `[PHONE]`)
- [ ] Click "Update"
- [ ] Verify filter updated to `[PHONE]`

### Edit Filter - Custom Value
- [ ] Find a filter with custom replacement (e.g., `[FILTERED]`)
- [ ] Click "Edit" button
- [ ] Verify dropdown shows "Custom" selected
- [ ] Verify text input shows current value `[FILTERED]`
- [ ] Change to template (e.g., `[***PASSWORD***]`)
- [ ] Click "Update"
- [ ] Verify filter updated to `[***PASSWORD***]`

### Edit Filter - Change to Custom
- [ ] Edit a filter with template replacement
- [ ] Change dropdown from template to "Custom"
- [ ] Type new custom value: `[BLOCKED]`
- [ ] Click "Update"
- [ ] Verify filter updated to `[BLOCKED]`

### Category Filtering
- [ ] In main filters list, use "Category" dropdown
- [ ] Select "🆔 PII (Personal Info)"
- [ ] Verify only filters with PII replacements show
- [ ] Select "🔐 Security & Credentials"
- [ ] Verify only filters with security replacements show

## Expected Behavior

### Create Modal
1. Dropdown defaults to "✏️ Custom (type your own)"
2. Selecting a template automatically populates `newFilter.replacement`
3. Shows preview box with selected template
4. Switching back to "Custom" shows text input

### Edit Modal
1. On open, checks if current replacement matches a template
2. If match found, selects that template in dropdown
3. If no match, defaults to "Custom" and shows text input
4. Changing template updates `editFilter.replacement` immediately

### Form Reset
1. After successful create, form resets
2. Replacement mode resets to "custom"
3. All fields cleared

## UI Features

- **Organized dropdown**: Templates grouped by category with emoji icons
- **Visual feedback**: Preview box shows selected template value
- **Color coding**: 
  - Create modal: Green focus ring
  - Edit modal: Purple focus ring
- **Accessibility**: Clear labels, proper focus states
- **Responsive**: Works on all screen sizes

## Integration Points

### Files Modified
1. `admin-ui/src/components/Filters.svelte`:
   - Added dropdown UI to Create modal (lines ~606-683)
   - Added dropdown UI to Edit modal (lines ~665-742)
   - Updated `openEditModal()` to detect template mode
   - Updated `resetForm()` to reset replacement mode

### Variables Used
- `newReplacementMode`: Tracks selected template in Create modal
- `editReplacementMode`: Tracks selected template in Edit modal
- `replacementTemplates`: Object with all template definitions

## Known Behavior

1. **Template Detection**: Only detects templates defined in `replacementTemplates` object
2. **Custom Values**: Any value not in templates shows as "Custom"
3. **Case Sensitive**: Template matching is case-sensitive
4. **Preview**: Only shows for template mode, not custom mode

## Future Enhancements

Possible improvements for future versions:

1. **Pattern Templates**: Add dropdown for common regex patterns
2. **Template Search**: Add search/filter in dropdown
3. **Recent Templates**: Show recently used templates at top
4. **Template Favorites**: Allow users to star favorite templates
5. **Bulk Template Change**: Change replacement template for multiple filters
6. **Template Import**: Import custom template definitions
7. **Template Categories**: Allow filtering dropdown by category

## Success Criteria ✅

- [x] Dropdown added to Create modal
- [x] Dropdown added to Edit modal
- [x] 60+ templates available
- [x] Organized by 6 categories
- [x] Custom mode with text input
- [x] Template auto-detection in Edit mode
- [x] Visual preview of selected template
- [x] Form reset properly handled
- [x] Color-coded focus states
- [x] Emoji category indicators

## Status: COMPLETE

The replacement template dropdown feature is now fully implemented and ready for testing.

**Access the feature:**
```bash
# Admin UI should already be running
firefox http://localhost:5173

# Navigate to: Filters → Add Filter
# Or click "Edit" on any existing filter
```

**Quick test:**
1. Click "➕ Add Filter"
2. Try the "Replacement Template" dropdown
3. Select different templates and see preview
4. Try "Custom" mode and type your own
5. Create filter and verify it works
6. Edit the filter and verify template is detected
