# SMTP Configuration Tutorial

This tutorial explains how to configure the SMTP settings for the email service in `el-campeon-web`.

## Required Environment Variables

Add the following variables to your `.env` file:

```env
SMTP_HOST=your-smtp-host
SMTP_PORT=your-smtp-port
SMTP_USER=your-smtp-username
SMTP_PASSWORD=your-smtp-password
SMTP_FROM_EMAIL=your-sender-email@example.com
```

---

## 1. Gmail Configuration (Recommended for Development)

To use Gmail, you **must** use an "App Password" if you have 2FA enabled (which is mandatory for Google).

1. Go to your [Google Account Settings](https://myaccount.google.com/).
2. Navigate to **Security**.
3. Under "Signing in to Google," select **2-Step Verification** and follow the steps to enable it.
4. Go back to Security and search for **App passwords**.
5. Select **Mail** and **Other (Custom name)**, name it "El Campeon Web".
6. Copy the 16-character password generated.

**Set these in your `.env`:**
```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-16-character-app-password
SMTP_FROM_EMAIL=your-email@gmail.com
```

---

## 2. Outlook / Hotmail Configuration

1. Go to your Outlook/Hotmail account.
2. Enable 2FA and generate an **App Password** if required.

**Set these in your `.env`:**
```env
SMTP_HOST=smtp.office365.com
SMTP_PORT=587
SMTP_USER=your-email@outlook.com
SMTP_PASSWORD=your-password-or-app-password
SMTP_FROM_EMAIL=your-email@outlook.com
```

---

## 3. SendGrid Configuration

If you prefer a transactional email service:

1. Create a SendGrid account.
2. Go to **Settings > API Keys** and create a new key with "Full Access" or "Mail Send" permissions.
3. Verify your **Sender Identity**.

**Set these in your `.env`:**
```env
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USER=apikey
SMTP_PASSWORD=your-sendgrid-api-key
SMTP_FROM_EMAIL=your-verified-sender@example.com
```

---

## 4. Mailtrap (For Testing)

Mailtrap is excellent for development as it catches all outgoing emails without actually sending them to the recipients.

1. Sign up at [Mailtrap.io](https://mailtrap.io/).
2. Go to **Email Testing > Inboxes**.
3. Click on your Inbox and select "Go" or "SMTP" integrations to see your credentials.

**Set these in your `.env`:**
```env
SMTP_HOST=sandbox.smtp.mailtrap.io
SMTP_PORT=2525
SMTP_USER=your-mailtrap-user
SMTP_PASSWORD=your-mailtrap-password
SMTP_FROM_EMAIL=test@elcampeon.com
```

---

## Troubleshooting

- **Port 587 vs 465**: This project uses port 587 (STARTTLS) by default. If your provider requires SSL/TLS on port 465, ensure the firewall allows it.
- **Authentication Failed**: Double-check that you are using an **App Password** for Gmail/Outlook, not your main account password.
- **Connection Timed Out**: Verify that your server/local machine allows outgoing connections on the specified SMTP port.
