package nagios

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccService_basic(t *testing.T) {
	serviceServiceName := "tf_" + acctest.RandString((10))
	serviceHostName := "localhost"
	serviceDescription := "tf_" + acctest.RandString(5)
	// serviceCheckCommand := "check_ping\\3000,80%\\5000,100%"
	serviceCheckCommand := "check_http"
	serviceMaxCheckAttempts := "2"
	serviceCheckInterval := "5"
	serviceRetryInterval := "5"
	serviceCheckPeriod := "24x7"
	serviceNotificationInterval := "10"
	serviceNotificationPeriod := "24x7"
	serviceContacts := "nagiosadmin"
	serviceTemplates := "generic-service"
	rName := "nagios_service.service"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceResourceBasic(serviceServiceName, serviceHostName, serviceDescription, serviceCheckCommand, serviceMaxCheckAttempts, serviceCheckInterval, serviceRetryInterval, serviceCheckPeriod, serviceNotificationInterval, serviceNotificationPeriod, serviceContacts, serviceTemplates),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(rName),
				),
			},
		},
	})
}

func TestAccServiceCreateAfterManualDestroy(t *testing.T) {
	var service = &Service{}
	serviceServiceName := "camd_" + acctest.RandString(10)
	serviceHostName := "localhost"
	serviceDescription := "This is a description"
	serviceCheckCommand := "check_http"
	serviceMaxCheckAttempts := "2"
	serviceCheckInterval := "5"
	serviceRetryInterval := "5"
	serviceCheckPeriod := "24x7"
	serviceNotificationInterval := "10"
	serviceNotificationPeriod := "24x7"
	serviceContacts := "nagiosadmin"
	serviceTemplates := "generic-service"
	rName := "nagios_service.service"

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckServiceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceResourceBasic(serviceServiceName, serviceHostName, serviceDescription, serviceCheckCommand, serviceMaxCheckAttempts, serviceCheckInterval, serviceRetryInterval, serviceCheckPeriod, serviceNotificationInterval, serviceNotificationPeriod, serviceContacts, serviceTemplates),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(rName),
					testAccCheckServiceFetch(rName, service),
				),
			},
			{
				PreConfig: func() {
					client := testAccProvider.Meta().(*Client)

					_, err := client.deleteService(mapArrayToString(service.HostName), service.Description)
					if err != nil {
						t.Fatal(err)
					}
				},
				Config: testAccServiceResourceBasic(serviceServiceName, serviceHostName, serviceDescription, serviceCheckCommand, serviceMaxCheckAttempts, serviceCheckInterval, serviceRetryInterval, serviceCheckPeriod, serviceNotificationInterval, serviceNotificationPeriod, serviceContacts, serviceTemplates),
				Check:  testAccCheckServiceExists(rName),
			},
		},
	})
}

func TestAccServiceUpdateName(t *testing.T) {
	firstServiceName := "tf_" + acctest.RandString(10)
	secondServiceName := "tf_" + acctest.RandString(10)
	serviceHostName := "localhost"
	serviceDescription := "tf_" + acctest.RandString(50)
	serviceCheckCommand := "check_ping!3000.0!80%!5000.0!100%!!!!"
	serviceMaxCheckAttempts := "2"
	serviceCheckInterval := "5"
	serviceRetryInterval := "5"
	serviceCheckPeriod := "24x7"
	serviceNotificationInterval := "10"
	serviceNotificationPeriod := "24x7"
	serviceContacts := "nagiosadmin"
	serviceTemplates := "generic-service"
	rName := "nagios_service.service"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccServiceResourceBasic(firstServiceName, serviceHostName, serviceDescription, serviceCheckCommand,
					serviceMaxCheckAttempts, serviceCheckInterval, serviceRetryInterval, serviceCheckPeriod, serviceNotificationInterval,
					serviceNotificationPeriod, serviceContacts, serviceTemplates),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(rName),
					resource.TestCheckResourceAttr(rName, "service_name", firstServiceName),
				),
			},
			{
				Config: testAccServiceResourceBasic(secondServiceName, serviceHostName, serviceDescription, serviceCheckCommand,
					serviceMaxCheckAttempts, serviceCheckInterval, serviceRetryInterval, serviceCheckPeriod, serviceNotificationInterval,
					serviceNotificationPeriod, serviceContacts, serviceTemplates),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceExists(rName),
					resource.TestCheckResourceAttr(rName, "service_name", secondServiceName),
				),
			},
		},
	})
}

func testAccServiceResourceBasic(serviceName, hostName, description, checkCommand, maxCheckAttempts, checkInterval, retryInterval, checkPeriod, notificationInterval, notificationPeriod, contacts, templates string) string {
	return fmt.Sprintf(`
resource "nagios_service" "service" {
	service_name			= "%s"
	host_name				= [
		"%s"
	]
	description				= "%s"
	check_command			= "%s"
	max_check_attempts		= "%s"
	check_interval			= "%s"
	retry_interval			= "%s"
	check_period			= "%s"
	notification_interval	= "%s"
	notification_period		= "%s"
	contacts				= [
		"%s"
	]
	templates				= [
		"%s"
	]
	free_variables 			= {
		"_test" = "TestVar123"
	}
}
	`, serviceName, hostName, description, checkCommand, maxCheckAttempts, checkInterval, retryInterval, checkPeriod, notificationInterval, notificationPeriod, contacts, templates)
}

func testAccCheckServiceDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "nagios_service" {
				continue
			}

			// Get the name from the state and check if it still exists
			name := rs.Primary.Attributes["service_name"]

			conn := testAccProvider.Meta().(*Client)

			service, _ := conn.getService(name)
			if service.ServiceName != "" {
				return fmt.Errorf("Service %s still exists", name)
			}
		}

		return nil
	}
}

func testAccCheckServiceExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getServiceFromState(s, resourceName)
		if err != nil {
			return err
		}

		return nil
	}
}

func getServiceFromState(s *terraform.State, rName string) (*Service, error) {
	nagiosClient := testAccProvider.Meta().(*Client)
	rs, ok := s.RootModule().Resources[rName]
	if !ok {
		return nil, fmt.Errorf("service not found: %s", rName)
	}

	name := rs.Primary.Attributes["service_name"]

	service, err := nagiosClient.getService(name)

	if err != nil {
		return nil, fmt.Errorf("error getting service with name %s: %s", name, err)
	}

	return service, nil
}

func testAccCheckServiceFetch(rName string, service *Service) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		returnedService, err := getServiceFromState(s, rName)
		if err != nil {
			return err
		}

		service.ServiceName = returnedService.ServiceName
		service.HostName = returnedService.HostName
		service.Description = returnedService.Description
		service.CheckCommand = returnedService.CheckCommand
		service.MaxCheckAttempts = returnedService.MaxCheckAttempts
		service.CheckInterval = returnedService.CheckInterval
		service.RetryInterval = returnedService.RetryInterval
		service.CheckPeriod = returnedService.CheckPeriod
		service.NotificationInterval = returnedService.NotificationInterval
		service.NotificationPeriod = returnedService.NotificationPeriod
		service.Contacts = returnedService.Contacts

		if returnedService.Templates != nil {
			service.Templates = returnedService.Templates
		}

		if returnedService.IsVolatile != "" {
			service.IsVolatile = returnedService.IsVolatile
		}

		if returnedService.InitialState != "" {
			service.InitialState = returnedService.InitialState
		}

		if returnedService.ActiveChecksEnabled != "" {
			service.ActiveChecksEnabled = returnedService.ActiveChecksEnabled
		}

		if returnedService.PassiveChecksEnabled != "" {
			service.PassiveChecksEnabled = returnedService.PassiveChecksEnabled
		}

		if returnedService.ObsessOverService != "" {
			service.ObsessOverService = returnedService.ObsessOverService
		}

		if returnedService.CheckFreshness != "" {
			service.CheckFreshness = returnedService.CheckFreshness
		}

		if returnedService.FreshnessThreshold != "" {
			service.FreshnessThreshold = returnedService.FreshnessThreshold
		}

		if returnedService.EventHandler != "" {
			service.EventHandler = returnedService.EventHandler
		}

		if returnedService.EventHandlerEnabled != "" {
			service.EventHandlerEnabled = returnedService.EventHandlerEnabled
		}

		if returnedService.LowFlapThreshold != "" {
			service.LowFlapThreshold = returnedService.LowFlapThreshold
		}

		if returnedService.HighFlapThreshold != "" {
			service.HighFlapThreshold = returnedService.HighFlapThreshold
		}

		if returnedService.FlapDetectionEnabled != "" {
			service.FlapDetectionEnabled = returnedService.FlapDetectionEnabled
		}

		if returnedService.FlapDetectionOptions != nil {
			service.FlapDetectionOptions = returnedService.FlapDetectionOptions
		}

		if returnedService.ProcessPerfData != "" {
			service.ProcessPerfData = returnedService.ProcessPerfData
		}

		if returnedService.RetainStatusInformation != "" {
			service.RetainStatusInformation = returnedService.RetainStatusInformation
		}

		if returnedService.RetainNonStatusInformation != "" {
			service.RetainNonStatusInformation = returnedService.RetainNonStatusInformation
		}

		if returnedService.FirstNotificationDelay != "" {
			service.FirstNotificationDelay = returnedService.FirstNotificationDelay
		}

		if returnedService.NotificationOptions != nil {
			service.NotificationOptions = returnedService.NotificationOptions
		}

		if returnedService.NotificationsEnabled != "" {
			service.NotificationsEnabled = returnedService.NotificationsEnabled
		}

		if returnedService.ContactGroups != nil {
			service.ContactGroups = returnedService.ContactGroups
		}

		if returnedService.Notes != "" {
			service.Notes = returnedService.Notes
		}

		if returnedService.NotesURL != "" {
			service.NotesURL = returnedService.NotesURL
		}

		if returnedService.ActionURL != "" {
			service.ActionURL = returnedService.ActionURL
		}

		if returnedService.IconImage != "" {
			service.IconImage = returnedService.IconImage
		}

		if returnedService.IconImageAlt != "" {
			service.IconImageAlt = returnedService.IconImageAlt
		}

		if returnedService.FreeVariables != nil {
			service.FreeVariables = returnedService.FreeVariables
		}

		return nil
	}
}
