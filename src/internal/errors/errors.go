package errsx

// Example usage:
//
//	app.Post("/register", func(c *http.Ctx) error {
//			// Define a struct for the request body
//			type User struct {
//				Username string `json:"username"`
//				Password string `json:"password"`
//				Email    string `json:"email"`
//			}
//			// simple validation for example purposes
//
//			var user User
//			// Parse the JSON body
//			if err := c.BodyParser(&user); err != nil {
//				return InvalidDataErr(err)
//			}
//			// Validate the data
//
//			errs errsx.ErrorMap
//			if len(user.Username) < 4 {
//				errs.Set("username", "Username must be at least 4 characters")
//			}
//			if len(user.Password) < 8 {
//				errs.Set("password", "Password must be at least 8 characters")
//			}
//			if len(user.Email) == 0 {
//				 errs.Set("email", "Email is required")
//			}
//			// Check if there were any errors
//			if errs != nil {
//				// Return the errors as a JSON response
//				return c.Status(http.StatusUnprocessableEntity).JSON(ValidationErr(errors))
//			}
//			// Continue with user registration process...
//			return c.SendStatus(http.StatusOK)
//	})
func ValidationErr(msg string, errors ErrorMap) error {
	return NewDesktopCleanerErrBuilder().
		WithCode(InvalidData).
		WithMsg(msg).
		WithDetails(NewDesktopCleanerErrDetails(errors))
}

func GenericErr(msg string, err error) error {
	return NewDesktopCleanerErrBuilder().
		WithCode(Internal).
		WithMsg(msg).
		WithCause(err)
}

func UnauthorizedErr(err error) error {
	return NewDesktopCleanerErrBuilder().
		WithCode(Unauthenticated).
		WithMsg("Unauthorized").
		WithCause(err)
}

func NotFoundErr(err error) error {
	return NewDesktopCleanerErrBuilder().
		WithCode(NotFound).
		WithMsg("Not Found").
		WithCause(err)
}
