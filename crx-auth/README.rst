===================================================
Chrome Extension for Getting Authentication Cookies
===================================================

This Chrome extension helps you get auth cookies of `Google Photos`_ for
gphotosuploader_

1. Install the extension in `Developer mode`_.
2. Log in to https://photos.google.com/
3. Click the icon of this extension. A popup windows will show up. Copy the
   content and save it in *auth.json*. The *userId* at the end of json file is
   blank.
4. See `this gist`_ to get your Google account/user ID. Fill the *userId* in the
   *auth.json* with the number.

This extension is released in public domain, see UNLICENSE_.

.. _Google Photos: https://photos.google.com/
.. _gphotosuploader: https://github.com/simonedegiacomi/gphotosuploader
.. _Developer mode: https://developer.chrome.com/extensions/getstarted#manifest
.. _this gist: https://gist.github.com/msafi/b1cb05bfab5b897c57e6
.. _UNLICENSE: https://unlicense.org/
