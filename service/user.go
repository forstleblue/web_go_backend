package service

import (
	"log"

	"github.com/unirep/ur-local-web/app/models/emailing"
	"github.com/unirep/ur-local-web/app/models/platform"
	"github.com/unirep/ur-local-web/app/models/user"
)

//User user service
type User struct {
	URLocalURL string
}

//LinkToPlatform inserts a profile on behalf of a platform, normally through the API
func (service *User) LinkToPlatform(p *user.Profile, plat *platform.Platform) error {
	id, err := user.InsertProfile(p)
	if err != nil {
		return err
	}
	p.ProfileID = id

	// send email
	emailing.SendNewPlatformProfileEmail(p, plat, service.URLocalURL)

	return nil
}

//RegistrationViaPlatform Registers a user plus initial platform profile at the same time
func (service *User) RegistrationViaPlatform(u *user.User, p *user.Profile, plat *platform.Platform) error {
	_, err := u.InitialRegister()
	if err != nil {
		return err
	}
	p.User = *u
	_, err = user.InsertProfile(p)

	// insert default customer profile
	service.createDefaultProfile(u)

	// send email
	emailing.SendRegistrationViaPlatformEmail(u, p, plat, service.URLocalURL)

	return nil
}

func (service *User) createDefaultProfile(usr *user.User) {
	service.createDefaultProfileWPhoto(usr, "")
}

func (service *User) createDefaultProfileWPhoto(usr *user.User, photoURL string) {
	profile := &user.Profile{}
	profile.User = *usr
	profile.ServiceCategory = 0
	profile.ProfileType = "b"
	profile.Description = "Default customer profile for buying goods and services."
	profile.Heading = "Customer"
	if photoURL != "" {
		profile.PhotoURL = photoURL
	}

	_, err := user.InsertProfile(profile)
	if err != nil {
		log.Printf("Error inserting default profile {%s}: %s\n", profile.String(), err.Error())
		// we should have a way of logging tasks like these that aren't created automatically but should be.
		// this shouldn't fail the registration, as the user was created successfully, but need to decide what we should do
		// otherwise, we need to setup transactions so that if one fails, they all get rolled back.
	}
}
