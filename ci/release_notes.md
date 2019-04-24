# Improvements

Uses the vaultkv library (used in safe) to automagically detect the KV
backend version and handle it seamlessly. It can even use a v1 and v2
backend in the same run! No more needing to give an environment
variable. This also fixes bugs where the vault mount had a slash in it.
Also, the old way was breaking most of the time because Safe doesn't
actually write an `api_version` key to .svtoken, and spruce was checking
for it.

# Acknowledgements

Thanks to @thomasmitchell for the update and fixes!
