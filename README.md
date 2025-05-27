# EXPO-Outlook-BookingHandler
Program to fetch booking from EXPO and Outlook then compare if the outlook has booked the same resource and then email the booker about their overbooking.


## env variables


Create on windows using:

```powershell
$env:NAME-OF-ENV = "Value of env"
```

EXPO_TOKEN = "Your EXPO auth token" 
EXPO_URL = "https://booking.yourdomain.com" The base URL to your EXPO booking system
TZ = "Europe/Stockholm" The timezone to use for the booking dates
SENDGRID_APIKEY = "Your SendGrid API Key"