package authentication

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func ServeTestAuth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(
		`    <!DOCTYPE html>
    <html>
    <head>
        <meta charset="UTF-8">
        <title>Sample FirebaseUI App</title>
        <script src="https://cdn.firebase.com/libs/firebaseui/2.5.1/firebaseui.js"></script>
        <link type="text/css" rel="stylesheet" href="https://cdn.firebase.com/libs/firebaseui/2.5.1/firebaseui.css" />

        <script src="https://www.gstatic.com/firebasejs/4.9.1/firebase.js"></script>
        <script>
            var config = {
                apiKey: "AIzaSyCqNlOLHg3FiCYCu2hkz3djXs8Tmw4DKpw",
                authDomain: "teddycare-193910.firebaseapp.com",
                databaseURL: "https://teddycare-193910.firebaseio.com",
                projectId: "teddycare-193910",
                storageBucket: "teddycare-193910.appspot.com",
                messagingSenderId: "36000389039"
            };
            firebase.initializeApp(config);

            // Initialize the FirebaseUI Widget using Firebase.
            var ui = new firebaseui.auth.AuthUI(firebase.auth());

            var uiConfig = {
                callbacks: {
                    signInSuccess: function(currentUser, credential, redirectUrl) {
                        firebase.auth().currentUser.getIdToken(/* forceRefresh */ true).then(function(idToken) {
                            console.log(idToken)
                            // Send token to your backend via HTTPS
                            // ...
                        }).catch(function(error) {
                            console.log(error)
                        });

                    },
                    uiShown: function() {
                        // The widget is rendered.
                        // Hide the loader.
                        document.getElementById('loader').style.display = 'none';
                    },
                    signInFailure: function(error) {
                        // For merge conflicts, the error.code will be
                        // 'firebaseui/anonymous-upgrade-merge-conflict'.
                        if (error.code != 'firebaseui/anonymous-upgrade-merge-conflict') {
                            return Promise.resolve();
                        }
                        // The credential the user tried to sign in with.
                        var cred = error.credential;
                        // If using Firebase Realtime Database. The anonymous user data has to be
                        // copied to the non-anonymous user.
                        var app = firebase.app();
                        // Save anonymous user data first.
                        return app.database().ref('users/' + firebase.auth().currentUser.uid)
                                .once('value')
                                .then(function(snapshot) {
                                    data = snapshot.val();
                                    // This will trigger onAuthStateChanged listener which
                                    // could trigger a redirect to another page.
                                    // Ensure the upgrade flow is not interrupted by that callback
                                    // and that this is given enough time to complete before
                                    // redirection.
                                    return firebase.auth().signInWithCredential(cred);
                                })
                                .then(function(user) {
                                    // Original Anonymous Auth instance now has the new user.
                                    return app.database().ref('users/' + user.uid).set(data);
                                })
                                .then(function() {
                                    // Delete anonymnous user.
                                    return anonymousUser.delete();
                                }).then(function() {
                                    // Clear data in case a new user signs in, and the state change
                                    // triggers.
                                    data = null;
                                    // FirebaseUI will reset and the UI cleared when this promise
                                    // resolves.
                                    // signInSuccess will not run. Successful sign-in logic has to be
                                    // run explicitly.
                                    window.location.assign('<url-to-redirect-to-on-success>');
                                });

                    }
                },
                // Will use popup for IDP Providers sign-in flow instead of the default, redirect.
                signInFlow: 'redirect',
                signInSuccessUrl: '/test-auth-on-success',
                signInOptions: [
                    // Leave the lines as is for the providers you want to offer your users.
                    firebase.auth.GoogleAuthProvider.PROVIDER_ID,
                    firebase.auth.FacebookAuthProvider.PROVIDER_ID,
                    firebase.auth.TwitterAuthProvider.PROVIDER_ID,
                    firebase.auth.GithubAuthProvider.PROVIDER_ID,
                    firebase.auth.EmailAuthProvider.PROVIDER_ID,
                    firebase.auth.PhoneAuthProvider.PROVIDER_ID
                ],
                // Terms of service url.
                tosUrl: '<your-tos-url>'
            };
            // Temp variable to hold the anonymous user data if needed.
            var data = null;
            // Hold a reference to the anonymous current user.
            var anonymousUser = firebase.auth().currentUser;
            ui.start('#firebaseui-auth-container', uiConfig);

        </script>
    </head>
    <body>
    <h1>Welcome to My Awesome App</h1>
    <div id="firebaseui-auth-container"></div>
    <div id="loader">Loading...</div>
    </body>
    </html>
`))
}

func ServeTestAuthOnSuccess(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("authorization")

	payload, _ := ioutil.ReadAll(r.Body)

	infos := map[string]interface{}{
		"token":   token,
		"payload": payload,
	}
	writeJSON(w, infos, 200)
}

func writeJSON(w http.ResponseWriter, data interface{}, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	switch v := data.(type) {
	case []byte:
		w.Write(v)
	case string:
		w.Write([]byte(v))
	default:
		json.NewEncoder(w).Encode(data)
	}
}
