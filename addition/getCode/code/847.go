package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
)

// SmartContract Define the Smart Contract structure
type SmartContract struct {
}

type admission_criteria struct {
	InstituteID         string `json:"InstituteID"`
	InstituteName       string `json:"Institute_Name"`
	Stream              string `json:"Stream"`
	MaxSeatCount        string `json:"Max_Seat_Count"`
	MinimumAge          string `json:"Minimum_Age"`
	MinimumRankExam     string `json:"Minimum_Rank_Exam"`
	MinimumBoardPercent string `json:"Minimum_Board_Percent"`
	Extras              string `json:"Extras"`
}

type prospective_student struct {
	Name       string `json:"name"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Address    string `json:"address"`
	AadharNo   string `json:"aadharno"`
	DOB        string `json:"DOB"`
	Email      string `json:"email"`
	Gender     string `json:"gender"`
	Mob_No     string `json:"mob_number"`
	AppliedFor string `json:"applied_for"`
}

type application_pool struct {
	ApplicationID  string `json:"ApplicationID"`
	InstituteName  string `json:"InstituteName"`
	Name           string `json:"Name"`
	DOB            string `json:"DOB"`
	Gender         string `json:"Gender"`
	Email          string `json:"email"`
	Mob_No         string `json:"Mobile_number"`
	Aadhar_no      string `json:"Aadhar_Number"`
	Marksheet_10   string `json:"Marksheet_10"`
	Marksheet_12   string `json:"Marksheet_12"`
	EntranceResult string `json:"EntranceResult"`
	Achievements   string `json:"Achievements"`
	Username       string `json:"Username"`
	Password       string `json:"password"`
	Status         string `json:"status"`
}

func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("DESy_cc")

// Invoke :  Method for INVOKING smart contract
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))

	switch function {
	case "createAdmissionCriteria":
		return s.createAdmissionCriteria(APIstub, args)
	case "updateAdmissionCriteria":
		return s.updateAdmissionCriteria(APIstub, args)
	case "queryCriteria":
		return s.queryCriteria(APIstub, args)
	case "createCourse":
		return s.createCourse(APIstub, args)
	case "updateCourse":
		return s.updateCourse(APIstub, args)
	case "queryCourse":
		return s.queryCourse(APIstub, args)
	case "prospectiveStudent":
		return s.prospectiveStudent(APIstub, args)
	case "readPStudent":
		return s.readPStudent(APIstub, args)
	case "createApplication":
		return s.createApplication(APIstub, args)
	case "readApplication":
		return s.readApplication(APIstub, args)
	case "updateApplication":
		return s.updateApplication(APIstub, args)
	case "queryAllApplications":
		return s.queryAllApplications(APIstub, args)
	case "transferApplication1":
		return s.transferApplication1(APIstub, args)
	case "updateApplication1":
		return s.updateApplication1(APIstub, args)
	case "queryAllApplications1":
		return s.queryAllApplications1(APIstub, args)
	case "transferApplication2":
		return s.transferApplication2(APIstub, args)
	case "updateApplication2":
		return s.updateApplication2(APIstub, args)
	case "queryAllApplications2":
		return s.queryAllApplications2(APIstub, args)
	case "transferApplication3":
		return s.transferApplication3(APIstub, args)
	case "updateApplication3":
		return s.updateApplication3(APIstub, args)
	case "queryAllApplications3":
		return s.queryAllApplications3(APIstub, args)
	case "hostchangeApplication":
		return s.hostchangeApplication(APIstub, args)
	case "dropoutApplication":
		return s.dropoutApplication(APIstub, args)
	}
	return shim.Error("Invalid Smart Contract function name.")
}

//Create Admission Criterias
func (s *SmartContract) createAdmissionCriteria(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 8 {
		return shim.Error("Incorrect number of arguments. Expecting 8")
	}

	var criteria = admission_criteria{InstituteID: args[0], InstituteName: args[1], Stream: args[2], MaxSeatCount: args[3], MinimumAge: args[4], MinimumRankExam: args[5], MinimumBoardPercent: args[6], Extras: args[7]}

	criteriaAsBytes, _ := json.Marshal(criteria)
	APIstub.PutState(args[0], criteriaAsBytes)

	return shim.Success(criteriaAsBytes)
}

//Update Admission Criterias
func (s *SmartContract) updateAdmissionCriteria(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	criteriaAsBytes, err := APIstub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get Criteria: " + err.Error())
	} else if criteriaAsBytes == nil {
		fmt.Println("This Criteria does not exists: " + args[0])
		return shim.Error("This Criteria does not exists: " + args[0])
	}

	updated := admission_criteria{}

	json.Unmarshal(criteriaAsBytes, &updated)
	toBeChanged := args[1]

	if toBeChanged == "MaxSeatCount" {
		updated.MaxSeatCount = args[2]
		criteriaAsBytes, _ = json.Marshal(updated)
		APIstub.PutState(args[0], criteriaAsBytes)
		return shim.Success(criteriaAsBytes)
	} else if toBeChanged == "MinimumAge" {
		updated.MinimumAge = args[2]
		criteriaAsBytes, _ = json.Marshal(updated)
		APIstub.PutState(args[0], criteriaAsBytes)
		return shim.Success(criteriaAsBytes)
	} else if toBeChanged == "MinimumRankExam" {
		updated.MinimumRankExam = args[2]
		criteriaAsBytes, _ = json.Marshal(updated)
		APIstub.PutState(args[0], criteriaAsBytes)
		return shim.Success(criteriaAsBytes)
	} else if toBeChanged == "MinimumBoardPercent" {
		updated.MinimumBoardPercent = args[2]
		criteriaAsBytes, _ = json.Marshal(updated)
		APIstub.PutState(args[0], criteriaAsBytes)
		return shim.Success(criteriaAsBytes)
	} else if toBeChanged == "Extras" {
		updated.Extras = args[2]
		criteriaAsBytes, _ = json.Marshal(updated)
		APIstub.PutState(args[0], criteriaAsBytes)
		return shim.Success(criteriaAsBytes)
	}
	return shim.Error("Field to be Updated is not correct")
}

//Query Admission Criterias
func (s *SmartContract) queryCriteria(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	criteriaAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(criteriaAsBytes)
}

type course struct {
	CourseID        string `json:"CourseID"`
	InstituteName   string `json:"Institute_Name"`
	Stream          string `json:"Stream"`
	TotalLectures   string `json:"TotaleLectures"`
	TotalTutorials  string `json:"TotalTutorials"`
	TotalPracticals string `json:"TotalPracticals"`
	CourseCredits   string `json:"CourseCredits"`
	CourseProfessor string `json:"CourseProfessor"`
	CourseSem       string `json:"CourseSemester"`
	CourseSyllabus  string `json:"CourseSyllabus"` //Syllabus PDF IPFS Hash
}

func (s *SmartContract) createCourse(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 9 {
		return shim.Error("Incorrect number of arguments. Expecting 9")
	}

	x := 0
	var id string
	for x == 0 {
		numb := strconv.Itoa(1000 + rand.Intn(9999-1000))
		id = args[0] + numb

		exists, _ := APIstub.GetState(id)

		if exists != nil {
			x = 0
		} else {
			x = 1
		}
	}

	course := course{CourseID: id, InstituteName: args[0], Stream: args[1], TotalLectures: args[2], TotalTutorials: args[3], TotalPracticals: args[4], CourseCredits: args[5], CourseProfessor: args[6], CourseSem: args[7], CourseSyllabus: args[8]}
	courseAsBytes, _ := json.Marshal(course)
	APIstub.PutState(id, courseAsBytes)

	return shim.Success(courseAsBytes)
}

//Update Course
func (s *SmartContract) updateCourse(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	courseAsBytes, err := APIstub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get Course: " + err.Error())
	} else if courseAsBytes == nil {
		fmt.Println("This Course does not exists: " + args[0])
		return shim.Error("This Course does not exists: " + args[0])
	}

	updated := course{}

	json.Unmarshal(courseAsBytes, &updated)
	toBeChanged := args[1]

	if toBeChanged == "TotalLectures" {
		updated.TotalLectures = args[2]
		courseAsBytes, _ = json.Marshal(updated)
		APIstub.PutState(args[0], courseAsBytes)
		return shim.Success(courseAsBytes)
	} else if toBeChanged == "TotalTutorials" {
		updated.TotalTutorials = args[2]
		courseAsBytes, _ = json.Marshal(updated)
		APIstub.PutState(args[0], courseAsBytes)
		return shim.Success(courseAsBytes)
	} else if toBeChanged == "TotalPracticals" {
		updated.TotalPracticals = args[2]
		courseAsBytes, _ = json.Marshal(updated)
		APIstub.PutState(args[0], courseAsBytes)
		return shim.Success(courseAsBytes)
	} else if toBeChanged == "CourseCredits" {
		updated.CourseCredits = args[2]
		courseAsBytes, _ = json.Marshal(updated)
		APIstub.PutState(args[0], courseAsBytes)
		return shim.Success(courseAsBytes)
	} else if toBeChanged == "CourseProfessor" {
		updated.CourseProfessor = args[2]
		courseAsBytes, _ = json.Marshal(updated)
		APIstub.PutState(args[0], courseAsBytes)
		return shim.Success(courseAsBytes)
	} else if toBeChanged == "CourseSem" {
		updated.CourseSem = args[2]
		courseAsBytes, _ = json.Marshal(updated)
		APIstub.PutState(args[0], courseAsBytes)
		return shim.Success(courseAsBytes)
	} else if toBeChanged == "CourseSyllabus" {
		updated.CourseSyllabus = args[2]
		courseAsBytes, _ = json.Marshal(updated)
		APIstub.PutState(args[0], courseAsBytes)
		return shim.Success(courseAsBytes)
	}
	return shim.Error("Field to be Updated is not correct")
}

//Query Course
func (s *SmartContract) queryCourse(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	courseAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(courseAsBytes)
}

//Onboarding Prospective Applicant
func (s *SmartContract) prospectiveStudent(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	transientMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("222222 -Error getting transient: " + err.Error())
	}
	transientAssetJSON, ok := transientMap["ProspectiveStudent"]
	if !ok {
		return shim.Error("Username must be a key in the transient map")
	}

	type assetPStudent struct {
		Name      string `json:"name"`
		DOB       string `json:"DOB"`
		Gender    string `json:"gender"`
		Email     string `json:"email"`
		Mob_No    string `json:"mob_number"`
		Aadhar_no string `json:"aadharno"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		Address   string `json:"address"`
	}

	var studentInput assetPStudent
	err = json.Unmarshal(transientAssetJSON, &studentInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(transientAssetJSON) + "Error is : " + err.Error())
	}

	// //check whether prospective student already exists
	studentAsBytes, err := APIstub.GetPrivateData("prospective_students", studentInput.Username)
	if err != nil {
		return shim.Error("Failed to get asset: " + err.Error())
	} else if studentAsBytes != nil {
		fmt.Println("Student already exists: " + studentInput.Username)
		return shim.Error("this student already exists: " + studentInput.Username)
	}

	student := &prospective_student{
		Name:       studentInput.Name,
		DOB:        studentInput.DOB,
		Gender:     studentInput.Gender,
		Email:      studentInput.Email,
		Mob_No:     studentInput.Mob_No,
		AadharNo:   studentInput.Aadhar_no,
		Username:   studentInput.Username,
		Password:   studentInput.Password,
		Address:    studentInput.Address,
		AppliedFor: "nil",
	}

	studentJSONasBytes, err := json.Marshal(student)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("prospective_students", studentInput.Username, studentJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(studentJSONasBytes)

}

//Query Prospective Student
func (s *SmartContract) readPStudent(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	studentbytes, err := APIstub.GetPrivateData("prospective_students", args[0])
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get private details for " + args[0] + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if studentbytes == nil {
		jsonResp := "{\"Error\":\"Student private details does not exist: " + args[0] + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(studentbytes)
}

//Create Application of Prospective Student
func (s *SmartContract) createApplication(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	transientMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("222222 -Error getting transient: " + err.Error())
	}
	transientAssetJSON, ok := transientMap["Application"]
	if !ok {
		return shim.Error("application must be a key in the transient map")
	}

	type assetStudent struct {
		ApplicationNumber string `json:"ApplicationNumber"`
		InstituteName     string `json:"InstituteName"`
		Name              string `json:"Name"`
		DOB               string `json:"DOB"`
		Gender            string `json:"Gender"`
		Email             string `json:"email"`
		Mob_No            string `json:"Mobile_number"`
		Aadhar_no         string `json:"Aadhar_Number"`
		Marksheet_10      string `json:"Marksheet_10"`
		Marksheet_12      string `json:"Marksheet_12"`
		EntranceResult    string `json:"EntranceResult"`
		Achievements      string `json:"Achievements"`
		Username          string `json:"Username"`
		Password          string `json:"password"`
	}

	var studentInput assetStudent
	err = json.Unmarshal(transientAssetJSON, &studentInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(transientAssetJSON) + "Error is : " + err.Error())
	}

	if len(studentInput.ApplicationNumber) == 0 {
		return shim.Error("ApplicationNumber field must be a non-empty string")
	}
	if len(studentInput.Name) == 0 {
		return shim.Error("Name field must be a non-empty string")
	}
	if len(studentInput.DOB) == 0 {
		return shim.Error("DOB field must be a non-empty string")
	}
	if len(studentInput.Email) == 0 {
		return shim.Error("Email field must be a non-empty string")
	}
	// if len(studentInput.Mob_No) == 0 {
	// 	return shim.Error("Mobile number entered should be 10 digits")
	// }
	if len(studentInput.Username) == 0 {
		return shim.Error("UserName field must be a non-empty string")
	}
	if len(studentInput.Password) == 0 {
		return shim.Error("Password field must be a non-empty string")
	}
	appId := studentInput.InstituteName + studentInput.ApplicationNumber
	// //check whether application already exists
	studentAsBytes, err := APIstub.GetPrivateData("application_pool", appId)
	if err != nil {
		return shim.Error("Failed to get Application: " + err.Error())
	} else if studentAsBytes != nil {
		fmt.Println("Application already exists: " + appId)
		return shim.Error("this Application already exists: " + appId)
	}

	//"{\"ApplicationID\":\"1111\", \"Name\":\"Aditya\", \"DOB\":\"01-02-03\", \"Gender\":\"Male\", \"email\":\"xyz@abc.com\", \"Mobile_number\":\"9876543210\", \"Aadhar_Number\":\"987698769876\", \"Marksheet_10\":\"csvqraev\", \"Marksheet_12\":\"Advregareitya\", \"EntranceResult\":\"vreagerb\", \"Achievements\":\"vreavteb\", \"Username\":\"Aditya\", \"Password\":\"Aditya\", \"School_10\":\"{\"schoolname\":\"XYZ\", \"schoolboard\":\"Aditya\", \"percentage\":\"80\",}\", \"School_12\":\"{\"schoolname\":\"XYZ\", \"schoolboard\":\"Aditya\", \"percentage\":\"80\"}\", \"Subject_12\":\"{\"Subject1\":\"{\"subjectname\":\"Maths\",\"subjectmarks\":\"70\"}\", \"Subject2\":\"{\"subjectname\":\"Physics\",\"subjectmarks\":\"70\"}\", \"Subject3\":\"{\"subjectname\":\"Chemistry\",\"subjectmarks\":\"70\"}\", \"Subject4\":\"{\"subjectname\":\"English\",\"subjectmarks\":\"70\"}\", \"Subject5\":\"{\"subjectname\":\"Computer\",\"subjectmarks\":\"70\"}\"}\", \"status\":\"true\"}"
	student := &application_pool{
		ApplicationID:  appId,
		InstituteName:  studentInput.InstituteName,
		Name:           studentInput.Name,
		DOB:            studentInput.DOB,
		Gender:         studentInput.Gender,
		Email:          studentInput.Email,
		Mob_No:         studentInput.Mob_No,
		Aadhar_no:      studentInput.Aadhar_no,
		Marksheet_10:   studentInput.Marksheet_10,
		Marksheet_12:   studentInput.Marksheet_12,
		EntranceResult: studentInput.EntranceResult,
		Achievements:   studentInput.Achievements,
		Username:       studentInput.Username,
		Password:       studentInput.Password,
		Status:         "Unverified",
	}

	studentJSONasBytes, err := json.Marshal(student)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("application_pool", appId, studentJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(studentJSONasBytes)
}

//Read Applications
func (s *SmartContract) readApplication(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	studentbytes, err := APIstub.GetPrivateData("application_pool", args[0])
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get private details for " + args[0] + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if studentbytes == nil {
		jsonResp := "{\"Error\":\"Application private details does not exist: " + args[0] + "\"}"
		return shim.Error(jsonResp)
	}

	return shim.Success(studentbytes)
}

func (s *SmartContract) queryAllApplications(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	instituteName := args[0]
	startKey := instituteName + "1"
	endKey := instituteName + "1000"

	resultsIterator, err := APIstub.GetPrivateDataByRange("application_pool", startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	results := []*application_pool{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var asset *application_pool
		err = json.Unmarshal(response.Value, &asset)
		if err != nil {
			return shim.Error(err.Error())
		}

		results = append(results, asset)
	}

	resultsJSON, _ := json.Marshal(results)

	return shim.Success(resultsJSON)
}

//Update Application of Prospective Student in Manageral
func (s *SmartContract) updateApplication(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	transientMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("222222 -Error getting transient: " + err.Error())
	}
	transientAssetJSON, ok := transientMap["Application"]
	if !ok {
		return shim.Error("application must be a key in the transient map")
	}

	type assetStudent struct {
		ApplicationNumber string `json:"ApplicationNumber"`
		InstituteName     string `json:"InstituteName"`
		Name              string `json:"Name"`
		DOB               string `json:"DOB"`
		Gender            string `json:"Gender"`
		Email             string `json:"email"`
		Mob_No            string `json:"Mobile_number"`
		Aadhar_no         string `json:"Aadhar_Number"`
		Marksheet_10      string `json:"Marksheet_10"`
		Marksheet_12      string `json:"Marksheet_12"`
		EntranceResult    string `json:"EntranceResult"`
		Achievements      string `json:"Achievements"`
		Username          string `json:"Username"`
		Password          string `json:"password"`
	}

	var studentInput assetStudent
	err = json.Unmarshal(transientAssetJSON, &studentInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(transientAssetJSON) + "Error is : " + err.Error())
	}

	if len(studentInput.ApplicationNumber) == 0 {
		return shim.Error("ApplicationID field must be a non-empty string")
	}
	if len(studentInput.Name) == 0 {
		return shim.Error("Name field must be a non-empty string")
	}
	if len(studentInput.DOB) == 0 {
		return shim.Error("DOB field must be a non-empty string")
	}
	if len(studentInput.Email) == 0 {
		return shim.Error("Email field must be a non-empty string")
	}
	// if len(studentInput.Mob_No) == 0 {
	// 	return shim.Error("Mobile number entered should be 10 digits")
	// }
	if len(studentInput.Username) == 0 {
		return shim.Error("UserName field must be a non-empty string")
	}
	if len(studentInput.Password) == 0 {
		return shim.Error("Password field must be a non-empty string")
	}
	appId := studentInput.InstituteName + studentInput.ApplicationNumber
	//"{\"ApplicationID\":\"1111\", \"Name\":\"Aditya\", \"DOB\":\"01-02-03\", \"Gender\":\"Male\", \"email\":\"xyz@abc.com\", \"Mobile_number\":\"9876543210\", \"Aadhar_Number\":\"987698769876\", \"Marksheet_10\":\"csvqraev\", \"Marksheet_12\":\"Advregareitya\", \"EntranceResult\":\"vreagerb\", \"Achievements\":\"vreavteb\", \"Username\":\"Aditya\", \"Password\":\"Aditya\", \"School_10\":\"{\"schoolname\":\"XYZ\", \"schoolboard\":\"Aditya\", \"percentage\":\"80\",}\", \"School_12\":\"{\"schoolname\":\"XYZ\", \"schoolboard\":\"Aditya\", \"percentage\":\"80\"}\", \"Subject_12\":\"{\"Subject1\":\"{\"subjectname\":\"Maths\",\"subjectmarks\":\"70\"}\", \"Subject2\":\"{\"subjectname\":\"Physics\",\"subjectmarks\":\"70\"}\", \"Subject3\":\"{\"subjectname\":\"Chemistry\",\"subjectmarks\":\"70\"}\", \"Subject4\":\"{\"subjectname\":\"English\",\"subjectmarks\":\"70\"}\", \"Subject5\":\"{\"subjectname\":\"Computer\",\"subjectmarks\":\"70\"}\"}\", \"status\":\"true\"}"
	student := &application_pool{
		ApplicationID:  appId,
		InstituteName:  studentInput.InstituteName,
		Name:           studentInput.Name,
		DOB:            studentInput.DOB,
		Gender:         studentInput.Gender,
		Email:          studentInput.Email,
		Mob_No:         studentInput.Mob_No,
		Aadhar_no:      studentInput.Aadhar_no,
		Marksheet_10:   studentInput.Marksheet_10,
		Marksheet_12:   studentInput.Marksheet_12,
		EntranceResult: studentInput.EntranceResult,
		Achievements:   studentInput.Achievements,
		Username:       studentInput.Username,
		Password:       studentInput.Password,
		Status:         "Verified",
	}

	studentJSONasBytes, err := json.Marshal(student)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("application_pool", appId, studentJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(studentJSONasBytes)
}

// **********************************INSTITUTE 1*******************************************

//Add Application of Prospective Student to the Institute 1
func (s *SmartContract) transferApplication1(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	transientMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("222222 -Error getting transient: " + err.Error())
	}
	transientAssetJSON, ok := transientMap["Application"]
	if !ok {
		return shim.Error("application must be a key in the transient map")
	}

	type assetStudent struct {
		ApplicationNumber string `json:"ApplicationNumber"`
		InstituteName     string `json:"InstituteName"`
		Name              string `json:"Name"`
		DOB               string `json:"DOB"`
		Gender            string `json:"Gender"`
		Email             string `json:"email"`
		Mob_No            string `json:"Mobile_number"`
		Aadhar_no         string `json:"Aadhar_Number"`
		Marksheet_10      string `json:"Marksheet_10"`
		Marksheet_12      string `json:"Marksheet_12"`
		EntranceResult    string `json:"EntranceResult"`
		Achievements      string `json:"Achievements"`
		Username          string `json:"Username"`
		Password          string `json:"password"`
	}

	var studentInput assetStudent
	err = json.Unmarshal(transientAssetJSON, &studentInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(transientAssetJSON) + "Error is : " + err.Error())
	}

	if len(studentInput.ApplicationNumber) == 0 {
		return shim.Error("ApplicationID field must be a non-empty string")
	}
	if len(studentInput.Name) == 0 {
		return shim.Error("Name field must be a non-empty string")
	}
	if len(studentInput.DOB) == 0 {
		return shim.Error("DOB field must be a non-empty string")
	}
	if len(studentInput.Email) == 0 {
		return shim.Error("Email field must be a non-empty string")
	}
	// if len(studentInput.Mob_No) == 0 {
	// 	return shim.Error("Mobile number entered should be 10 digits")
	// }
	if len(studentInput.Username) == 0 {
		return shim.Error("UserName field must be a non-empty string")
	}
	if len(studentInput.Password) == 0 {
		return shim.Error("Password field must be a non-empty string")
	}
	appId := studentInput.InstituteName + studentInput.ApplicationNumber
	// //check whether application already exists
	studentAsBytes, err := APIstub.GetPrivateData("institute1_students", appId)
	if err != nil {
		return shim.Error("Failed to get Application: " + err.Error())
	} else if studentAsBytes != nil {
		fmt.Println("Application already exists: " + appId)
		return shim.Error("this Application already exists: " + appId)
	}

	//"{\"ApplicationID\":\"1111\", \"Name\":\"Aditya\", \"DOB\":\"01-02-03\", \"Gender\":\"Male\", \"email\":\"xyz@abc.com\", \"Mobile_number\":\"9876543210\", \"Aadhar_Number\":\"987698769876\", \"Marksheet_10\":\"csvqraev\", \"Marksheet_12\":\"Advregareitya\", \"EntranceResult\":\"vreagerb\", \"Achievements\":\"vreavteb\", \"Username\":\"Aditya\", \"Password\":\"Aditya\", \"School_10\":\"{\"schoolname\":\"XYZ\", \"schoolboard\":\"Aditya\", \"percentage\":\"80\",}\", \"School_12\":\"{\"schoolname\":\"XYZ\", \"schoolboard\":\"Aditya\", \"percentage\":\"80\"}\", \"Subject_12\":\"{\"Subject1\":\"{\"subjectname\":\"Maths\",\"subjectmarks\":\"70\"}\", \"Subject2\":\"{\"subjectname\":\"Physics\",\"subjectmarks\":\"70\"}\", \"Subject3\":\"{\"subjectname\":\"Chemistry\",\"subjectmarks\":\"70\"}\", \"Subject4\":\"{\"subjectname\":\"English\",\"subjectmarks\":\"70\"}\", \"Subject5\":\"{\"subjectname\":\"Computer\",\"subjectmarks\":\"70\"}\"}\", \"status\":\"true\"}"
	student := &application_pool{
		ApplicationID:  appId,
		InstituteName:  studentInput.InstituteName,
		Name:           studentInput.Name,
		DOB:            studentInput.DOB,
		Gender:         studentInput.Gender,
		Email:          studentInput.Email,
		Mob_No:         studentInput.Mob_No,
		Aadhar_no:      studentInput.Aadhar_no,
		Marksheet_10:   studentInput.Marksheet_10,
		Marksheet_12:   studentInput.Marksheet_12,
		EntranceResult: studentInput.EntranceResult,
		Achievements:   studentInput.Achievements,
		Username:       studentInput.Username,
		Password:       studentInput.Password,
		Status:         "Awaited",
	}

	studentJSONasBytes, err := json.Marshal(student)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("institute1_students", appId, studentJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(studentJSONasBytes)
}

//Update Application of Prospective Student of the Institute 1
func (s *SmartContract) updateApplication1(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	transientMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("222222 -Error getting transient: " + err.Error())
	}
	transientAssetJSON, ok := transientMap["Application"]
	if !ok {
		return shim.Error("application must be a key in the transient map")
	}

	type assetStudent struct {
		ApplicationNumber string `json:"ApplicationNumber"`
		InstituteName     string `json:"InstituteName"`
		Name              string `json:"Name"`
		DOB               string `json:"DOB"`
		Gender            string `json:"Gender"`
		Email             string `json:"email"`
		Mob_No            string `json:"Mobile_number"`
		Aadhar_no         string `json:"Aadhar_Number"`
		Marksheet_10      string `json:"Marksheet_10"`
		Marksheet_12      string `json:"Marksheet_12"`
		EntranceResult    string `json:"EntranceResult"`
		Achievements      string `json:"Achievements"`
		Username          string `json:"Username"`
		Password          string `json:"password"`
	}

	var studentInput assetStudent
	err = json.Unmarshal(transientAssetJSON, &studentInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(transientAssetJSON) + "Error is : " + err.Error())
	}

	if len(studentInput.ApplicationNumber) == 0 {
		return shim.Error("ApplicationNumber field must be a non-empty string")
	}
	if len(studentInput.Name) == 0 {
		return shim.Error("Name field must be a non-empty string")
	}
	if len(studentInput.DOB) == 0 {
		return shim.Error("DOB field must be a non-empty string")
	}
	if len(studentInput.Email) == 0 {
		return shim.Error("Email field must be a non-empty string")
	}
	// if len(studentInput.Mob_No) == 0 {
	// 	return shim.Error("Mobile number entered should be 10 digits")
	// }
	if len(studentInput.Username) == 0 {
		return shim.Error("UserName field must be a non-empty string")
	}
	if len(studentInput.Password) == 0 {
		return shim.Error("Password field must be a non-empty string")
	}
	appId := studentInput.InstituteName + studentInput.ApplicationNumber
	// //check whether application already exists
	studentAsBytes, err := APIstub.GetPrivateData("institute1_students", appId)
	if err != nil {
		return shim.Error("Failed to get Application: " + err.Error())
	} else if studentAsBytes != nil {
		fmt.Println("Application already exists: " + appId)
		return shim.Error("this Application already exists: " + appId)
	}

	//"{\"ApplicationID\":\"1111\", \"Name\":\"Aditya\", \"DOB\":\"01-02-03\", \"Gender\":\"Male\", \"email\":\"xyz@abc.com\", \"Mobile_number\":\"9876543210\", \"Aadhar_Number\":\"987698769876\", \"Marksheet_10\":\"csvqraev\", \"Marksheet_12\":\"Advregareitya\", \"EntranceResult\":\"vreagerb\", \"Achievements\":\"vreavteb\", \"Username\":\"Aditya\", \"Password\":\"Aditya\", \"School_10\":\"{\"schoolname\":\"XYZ\", \"schoolboard\":\"Aditya\", \"percentage\":\"80\",}\", \"School_12\":\"{\"schoolname\":\"XYZ\", \"schoolboard\":\"Aditya\", \"percentage\":\"80\"}\", \"Subject_12\":\"{\"Subject1\":\"{\"subjectname\":\"Maths\",\"subjectmarks\":\"70\"}\", \"Subject2\":\"{\"subjectname\":\"Physics\",\"subjectmarks\":\"70\"}\", \"Subject3\":\"{\"subjectname\":\"Chemistry\",\"subjectmarks\":\"70\"}\", \"Subject4\":\"{\"subjectname\":\"English\",\"subjectmarks\":\"70\"}\", \"Subject5\":\"{\"subjectname\":\"Computer\",\"subjectmarks\":\"70\"}\"}\", \"status\":\"true\"}"
	student := &application_pool{
		ApplicationID:  appId,
		InstituteName:  studentInput.InstituteName,
		Name:           studentInput.Name,
		DOB:            studentInput.DOB,
		Gender:         studentInput.Gender,
		Email:          studentInput.Email,
		Mob_No:         studentInput.Mob_No,
		Aadhar_no:      studentInput.Aadhar_no,
		Marksheet_10:   studentInput.Marksheet_10,
		Marksheet_12:   studentInput.Marksheet_12,
		EntranceResult: studentInput.EntranceResult,
		Achievements:   studentInput.Achievements,
		Username:       studentInput.Username,
		Password:       studentInput.Password,
		Status:         "Accepted",
	}

	studentJSONasBytes, err := json.Marshal(student)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("institute1_students", appId, studentJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(studentJSONasBytes)
}

func (s *SmartContract) queryAllApplications1(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	instituteName := args[0]
	startKey := instituteName + "1"
	endKey := instituteName + "1000"

	resultsIterator, err := APIstub.GetPrivateDataByRange("institute1_students", startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	results := []*application_pool{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var asset *application_pool
		err = json.Unmarshal(response.Value, &asset)
		if err != nil {
			return shim.Error(err.Error())
		}

		results = append(results, asset)
	}

	resultsJSON, _ := json.Marshal(results)

	return shim.Success(resultsJSON)
}

// **********************************INSTITUTE 2*******************************************

//Add Application of Prospective Student to the Institute 2
func (s *SmartContract) transferApplication2(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	transientMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("222222 -Error getting transient: " + err.Error())
	}
	transientAssetJSON, ok := transientMap["Application"]
	if !ok {
		return shim.Error("application must be a key in the transient map")
	}

	type assetStudent struct {
		ApplicationNumber string `json:"ApplicationNumber"`
		InstituteName     string `json:"InstituteName"`
		Name              string `json:"Name"`
		DOB               string `json:"DOB"`
		Gender            string `json:"Gender"`
		Email             string `json:"email"`
		Mob_No            string `json:"Mobile_number"`
		Aadhar_no         string `json:"Aadhar_Number"`
		Marksheet_10      string `json:"Marksheet_10"`
		Marksheet_12      string `json:"Marksheet_12"`
		EntranceResult    string `json:"EntranceResult"`
		Achievements      string `json:"Achievements"`
		Username          string `json:"Username"`
		Password          string `json:"password"`
	}

	var studentInput assetStudent
	err = json.Unmarshal(transientAssetJSON, &studentInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(transientAssetJSON) + "Error is : " + err.Error())
	}

	if len(studentInput.ApplicationNumber) == 0 {
		return shim.Error("Applicat ionNumber field must be a non-empty string")
	}
	if len(studentInput.Name) == 0 {
		return shim.Error("Name field must be a non-empty string")
	}
	if len(studentInput.DOB) == 0 {
		return shim.Error("DOB field must be a non-empty string")
	}
	if len(studentInput.Email) == 0 {
		return shim.Error("Email field must be a non-empty string")
	}
	// if len(studentInput.Mob_No) == 0 {
	// 	return shim.Error("Mobile number entered should be 10 digits")
	// }
	if len(studentInput.Username) == 0 {
		return shim.Error("UserName field must be a non-empty string")
	}
	if len(studentInput.Password) == 0 {
		return shim.Error("Password field must be a non-empty string")
	}
	appId := studentInput.InstituteName + studentInput.ApplicationNumber
	// //check whether application already exists
	studentAsBytes, err := APIstub.GetPrivateData("institute2_students", appId)
	if err != nil {
		return shim.Error("Failed to get Application: " + err.Error())
	} else if studentAsBytes != nil {
		fmt.Println("Application already exists: " + appId)
		return shim.Error("this Application already exists: " + appId)
	}

	//"{\"ApplicationID\":\"1111\", \"Name\":\"Aditya\", \"DOB\":\"01-02-03\", \"Gender\":\"Male\", \"email\":\"xyz@abc.com\", \"Mobile_number\":\"9876543210\", \"Aadhar_Number\":\"987698769876\", \"Marksheet_10\":\"csvqraev\", \"Marksheet_12\":\"Advregareitya\", \"EntranceResult\":\"vreagerb\", \"Achievements\":\"vreavteb\", \"Username\":\"Aditya\", \"Password\":\"Aditya\", \"School_10\":\"{\"schoolname\":\"XYZ\", \"schoolboard\":\"Aditya\", \"percentage\":\"80\",}\", \"School_12\":\"{\"schoolname\":\"XYZ\", \"schoolboard\":\"Aditya\", \"percentage\":\"80\"}\", \"Subject_12\":\"{\"Subject1\":\"{\"subjectname\":\"Maths\",\"subjectmarks\":\"70\"}\", \"Subject2\":\"{\"subjectname\":\"Physics\",\"subjectmarks\":\"70\"}\", \"Subject3\":\"{\"subjectname\":\"Chemistry\",\"subjectmarks\":\"70\"}\", \"Subject4\":\"{\"subjectname\":\"English\",\"subjectmarks\":\"70\"}\", \"Subject5\":\"{\"subjectname\":\"Computer\",\"subjectmarks\":\"70\"}\"}\", \"status\":\"true\"}"
	student := &application_pool{
		ApplicationID:  appId,
		InstituteName:  studentInput.InstituteName,
		Name:           studentInput.Name,
		DOB:            studentInput.DOB,
		Gender:         studentInput.Gender,
		Email:          studentInput.Email,
		Mob_No:         studentInput.Mob_No,
		Aadhar_no:      studentInput.Aadhar_no,
		Marksheet_10:   studentInput.Marksheet_10,
		Marksheet_12:   studentInput.Marksheet_12,
		EntranceResult: studentInput.EntranceResult,
		Achievements:   studentInput.Achievements,
		Username:       studentInput.Username,
		Password:       studentInput.Password,
		Status:         "Awaited",
	}

	studentJSONasBytes, err := json.Marshal(student)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("institute2_students", appId, studentJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(studentJSONasBytes)
}

//Update Application of Prospective Student of the Institute 2
func (s *SmartContract) updateApplication2(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	transientMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("222222 -Error getting transient: " + err.Error())
	}
	transientAssetJSON, ok := transientMap["Application"]
	if !ok {
		return shim.Error("application must be a key in the transient map")
	}

	type assetStudent struct {
		ApplicationNumber string `json:"ApplicationNumber"`
		InstituteName     string `json:"InstituteName"`
		Name              string `json:"Name"`
		DOB               string `json:"DOB"`
		Gender            string `json:"Gender"`
		Email             string `json:"email"`
		Mob_No            string `json:"Mobile_number"`
		Aadhar_no         string `json:"Aadhar_Number"`
		Marksheet_10      string `json:"Marksheet_10"`
		Marksheet_12      string `json:"Marksheet_12"`
		EntranceResult    string `json:"EntranceResult"`
		Achievements      string `json:"Achievements"`
		Username          string `json:"Username"`
		Password          string `json:"password"`
	}

	var studentInput assetStudent
	err = json.Unmarshal(transientAssetJSON, &studentInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(transientAssetJSON) + "Error is : " + err.Error())
	}

	if len(studentInput.ApplicationNumber) == 0 {
		return shim.Error("ApplicationNumber field must be a non-empty string")
	}
	if len(studentInput.Name) == 0 {
		return shim.Error("Name field must be a non-empty string")
	}
	if len(studentInput.DOB) == 0 {
		return shim.Error("DOB field must be a non-empty string")
	}
	if len(studentInput.Email) == 0 {
		return shim.Error("Email field must be a non-empty string")
	}
	// if len(studentInput.Mob_No) == 0 {
	// 	return shim.Error("Mobile number entered should be 10 digits")
	// }
	if len(studentInput.Username) == 0 {
		return shim.Error("UserName field must be a non-empty string")
	}
	if len(studentInput.Password) == 0 {
		return shim.Error("Password field must be a non-empty string")
	}
	appId := studentInput.InstituteName + studentInput.ApplicationNumber
	// //check whether application already exists
	studentAsBytes, err := APIstub.GetPrivateData("institute2_students", appId)
	if err != nil {
		return shim.Error("Failed to get Application: " + err.Error())
	} else if studentAsBytes != nil {
		fmt.Println("Application already exists: " + appId)
		return shim.Error("this Application already exists: " + appId)
	}

	//"{\"ApplicationID\":\"1111\", \"Name\":\"Aditya\", \"DOB\":\"01-02-03\", \"Gender\":\"Male\", \"email\":\"xyz@abc.com\", \"Mobile_number\":\"9876543210\", \"Aadhar_Number\":\"987698769876\", \"Marksheet_10\":\"csvqraev\", \"Marksheet_12\":\"Advregareitya\", \"EntranceResult\":\"vreagerb\", \"Achievements\":\"vreavteb\", \"Username\":\"Aditya\", \"Password\":\"Aditya\", \"School_10\":\"{\"schoolname\":\"XYZ\", \"schoolboard\":\"Aditya\", \"percentage\":\"80\",}\", \"School_12\":\"{\"schoolname\":\"XYZ\", \"schoolboard\":\"Aditya\", \"percentage\":\"80\"}\", \"Subject_12\":\"{\"Subject1\":\"{\"subjectname\":\"Maths\",\"subjectmarks\":\"70\"}\", \"Subject2\":\"{\"subjectname\":\"Physics\",\"subjectmarks\":\"70\"}\", \"Subject3\":\"{\"subjectname\":\"Chemistry\",\"subjectmarks\":\"70\"}\", \"Subject4\":\"{\"subjectname\":\"English\",\"subjectmarks\":\"70\"}\", \"Subject5\":\"{\"subjectname\":\"Computer\",\"subjectmarks\":\"70\"}\"}\", \"status\":\"true\"}"
	student := &application_pool{
		ApplicationID:  appId,
		InstituteName:  studentInput.InstituteName,
		Name:           studentInput.Name,
		DOB:            studentInput.DOB,
		Gender:         studentInput.Gender,
		Email:          studentInput.Email,
		Mob_No:         studentInput.Mob_No,
		Aadhar_no:      studentInput.Aadhar_no,
		Marksheet_10:   studentInput.Marksheet_10,
		Marksheet_12:   studentInput.Marksheet_12,
		EntranceResult: studentInput.EntranceResult,
		Achievements:   studentInput.Achievements,
		Username:       studentInput.Username,
		Password:       studentInput.Password,
		Status:         "Accepted",
	}

	studentJSONasBytes, err := json.Marshal(student)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("institute2_students", appId, studentJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(studentJSONasBytes)
}

func (s *SmartContract) queryAllApplications2(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	instituteName := args[0]
	startKey := instituteName + "1"
	endKey := instituteName + "1000"

	resultsIterator, err := APIstub.GetPrivateDataByRange("institute2_students", startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	results := []*application_pool{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var asset *application_pool
		err = json.Unmarshal(response.Value, &asset)
		if err != nil {
			return shim.Error(err.Error())
		}

		results = append(results, asset)
	}

	resultsJSON, _ := json.Marshal(results)

	return shim.Success(resultsJSON)
}

// **********************************INSTITUTE 3*******************************************

//Add Application of Prospective Student to the Institute 3
func (s *SmartContract) transferApplication3(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	transientMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("222222 -Error getting transient: " + err.Error())
	}
	transientAssetJSON, ok := transientMap["Application"]
	if !ok {
		return shim.Error("application must be a key in the transient map")
	}

	type assetStudent struct {
		ApplicationNumber string `json:"ApplicationNumber"`
		InstituteName     string `json:"InstituteName"`
		Name              string `json:"Name"`
		DOB               string `json:"DOB"`
		Gender            string `json:"Gender"`
		Email             string `json:"email"`
		Mob_No            string `json:"Mobile_number"`
		Aadhar_no         string `json:"Aadhar_Number"`
		Marksheet_10      string `json:"Marksheet_10"`
		Marksheet_12      string `json:"Marksheet_12"`
		EntranceResult    string `json:"EntranceResult"`
		Achievements      string `json:"Achievements"`
		Username          string `json:"Username"`
		Password          string `json:"password"`
	}

	var studentInput assetStudent
	err = json.Unmarshal(transientAssetJSON, &studentInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(transientAssetJSON) + "Error is : " + err.Error())
	}

	if len(studentInput.ApplicationNumber) == 0 {
		return shim.Error("ApplicationID field must be a non-empty string")
	}
	if len(studentInput.Name) == 0 {
		return shim.Error("Name field must be a non-empty string")
	}
	if len(studentInput.DOB) == 0 {
		return shim.Error("DOB field must be a non-empty string")
	}
	if len(studentInput.Email) == 0 {
		return shim.Error("Email field must be a non-empty string")
	}
	// if len(studentInput.Mob_No) == 0 {
	// 	return shim.Error("Mobile number entered should be 10 digits")
	// }
	if len(studentInput.Username) == 0 {
		return shim.Error("UserName field must be a non-empty string")
	}
	if len(studentInput.Password) == 0 {
		return shim.Error("Password field must be a non-empty string")
	}
	appId := studentInput.InstituteName + studentInput.ApplicationNumber
	// //check whether application already exists
	studentAsBytes, err := APIstub.GetPrivateData("institute3_students", appId)
	if err != nil {
		return shim.Error("Failed to get Application: " + err.Error())
	} else if studentAsBytes != nil {
		fmt.Println("Application already exists: " + appId)
		return shim.Error("this Application already exists: " + appId)
	}

	//"{\"ApplicationID\":\"1111\", \"Name\":\"Aditya\", \"DOB\":\"01-02-03\", \"Gender\":\"Male\", \"email\":\"xyz@abc.com\", \"Mobile_number\":\"9876543210\", \"Aadhar_Number\":\"987698769876\", \"Marksheet_10\":\"csvqraev\", \"Marksheet_12\":\"Advregareitya\", \"EntranceResult\":\"vreagerb\", \"Achievements\":\"vreavteb\", \"Username\":\"Aditya\", \"Password\":\"Aditya\", \"School_10\":\"{\"schoolname\":\"XYZ\", \"schoolboard\":\"Aditya\", \"percentage\":\"80\",}\", \"School_12\":\"{\"schoolname\":\"XYZ\", \"schoolboard\":\"Aditya\", \"percentage\":\"80\"}\", \"Subject_12\":\"{\"Subject1\":\"{\"subjectname\":\"Maths\",\"subjectmarks\":\"70\"}\", \"Subject2\":\"{\"subjectname\":\"Physics\",\"subjectmarks\":\"70\"}\", \"Subject3\":\"{\"subjectname\":\"Chemistry\",\"subjectmarks\":\"70\"}\", \"Subject4\":\"{\"subjectname\":\"English\",\"subjectmarks\":\"70\"}\", \"Subject5\":\"{\"subjectname\":\"Computer\",\"subjectmarks\":\"70\"}\"}\", \"status\":\"true\"}"
	student := &application_pool{
		ApplicationID:  appId,
		InstituteName:  studentInput.InstituteName,
		Name:           studentInput.Name,
		DOB:            studentInput.DOB,
		Gender:         studentInput.Gender,
		Email:          studentInput.Email,
		Mob_No:         studentInput.Mob_No,
		Aadhar_no:      studentInput.Aadhar_no,
		Marksheet_10:   studentInput.Marksheet_10,
		Marksheet_12:   studentInput.Marksheet_12,
		EntranceResult: studentInput.EntranceResult,
		Achievements:   studentInput.Achievements,
		Username:       studentInput.Username,
		Password:       studentInput.Password,
		Status:         "Awaited",
	}

	studentJSONasBytes, err := json.Marshal(student)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("institute3_students", appId, studentJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(studentJSONasBytes)
}

//Update Application of Prospective Student of the Institute 3
func (s *SmartContract) updateApplication3(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	transientMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("222222 -Error getting transient: " + err.Error())
	}
	transientAssetJSON, ok := transientMap["Application"]
	if !ok {
		return shim.Error("application must be a key in the transient map")
	}

	type assetStudent struct {
		ApplicationNumber string `json:"ApplicationNumber"`
		InstituteName     string `json:"InstituteName"`
		Name              string `json:"Name"`
		DOB               string `json:"DOB"`
		Gender            string `json:"Gender"`
		Email             string `json:"email"`
		Mob_No            string `json:"Mobile_number"`
		Aadhar_no         string `json:"Aadhar_Number"`
		Marksheet_10      string `json:"Marksheet_10"`
		Marksheet_12      string `json:"Marksheet_12"`
		EntranceResult    string `json:"EntranceResult"`
		Achievements      string `json:"Achievements"`
		Username          string `json:"Username"`
		Password          string `json:"password"`
	}

	var studentInput assetStudent
	err = json.Unmarshal(transientAssetJSON, &studentInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(transientAssetJSON) + "Error is : " + err.Error())
	}

	if len(studentInput.ApplicationNumber) == 0 {
		return shim.Error("ApplicationNumber field must be a non-empty string")
	}
	if len(studentInput.Name) == 0 {
		return shim.Error("Name field must be a non-empty string")
	}
	if len(studentInput.DOB) == 0 {
		return shim.Error("DOB field must be a non-empty string")
	}
	if len(studentInput.Email) == 0 {
		return shim.Error("Email field must be a non-empty string")
	}
	// if len(studentInput.Mob_No) == 0 {
	// 	return shim.Error("Mobile number entered should be 10 digits")
	// }
	if len(studentInput.Username) == 0 {
		return shim.Error("UserName field must be a non-empty string")
	}
	if len(studentInput.Password) == 0 {
		return shim.Error("Password field must be a non-empty string")
	}
	appId := studentInput.InstituteName + studentInput.ApplicationNumber
	// //check whether application already exists
	studentAsBytes, err := APIstub.GetPrivateData("institute3_students", appId)
	if err != nil {
		return shim.Error("Failed to get Application: " + err.Error())
	} else if studentAsBytes != nil {
		fmt.Println("Application already exists: " + appId)
		return shim.Error("this Application already exists: " + appId)
	}

	//"{\"ApplicationID\":\"1111\", \"Name\":\"Aditya\", \"DOB\":\"01-02-03\", \"Gender\":\"Male\", \"email\":\"xyz@abc.com\", \"Mobile_number\":\"9876543210\", \"Aadhar_Number\":\"987698769876\", \"Marksheet_10\":\"csvqraev\", \"Marksheet_12\":\"Advregareitya\", \"EntranceResult\":\"vreagerb\", \"Achievements\":\"vreavteb\", \"Username\":\"Aditya\", \"Password\":\"Aditya\", \"School_10\":\"{\"schoolname\":\"XYZ\", \"schoolboard\":\"Aditya\", \"percentage\":\"80\",}\", \"School_12\":\"{\"schoolname\":\"XYZ\", \"schoolboard\":\"Aditya\", \"percentage\":\"80\"}\", \"Subject_12\":\"{\"Subject1\":\"{\"subjectname\":\"Maths\",\"subjectmarks\":\"70\"}\", \"Subject2\":\"{\"subjectname\":\"Physics\",\"subjectmarks\":\"70\"}\", \"Subject3\":\"{\"subjectname\":\"Chemistry\",\"subjectmarks\":\"70\"}\", \"Subject4\":\"{\"subjectname\":\"English\",\"subjectmarks\":\"70\"}\", \"Subject5\":\"{\"subjectname\":\"Computer\",\"subjectmarks\":\"70\"}\"}\", \"status\":\"true\"}"
	student := &application_pool{
		ApplicationID:  appId,
		InstituteName:  studentInput.InstituteName,
		Name:           studentInput.Name,
		DOB:            studentInput.DOB,
		Gender:         studentInput.Gender,
		Email:          studentInput.Email,
		Mob_No:         studentInput.Mob_No,
		Aadhar_no:      studentInput.Aadhar_no,
		Marksheet_10:   studentInput.Marksheet_10,
		Marksheet_12:   studentInput.Marksheet_12,
		EntranceResult: studentInput.EntranceResult,
		Achievements:   studentInput.Achievements,
		Username:       studentInput.Username,
		Password:       studentInput.Password,
		Status:         "Accepted",
	}

	studentJSONasBytes, err := json.Marshal(student)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("institute3_students", appId, studentJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(studentJSONasBytes)
}

func (s *SmartContract) queryAllApplications3(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	instituteName := args[0]
	startKey := instituteName + "1"
	endKey := instituteName + "1000"

	resultsIterator, err := APIstub.GetPrivateDataByRange("institute3_students", startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	results := []*application_pool{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		var asset *application_pool
		err = json.Unmarshal(response.Value, &asset)
		if err != nil {
			return shim.Error(err.Error())
		}

		results = append(results, asset)
	}

	resultsJSON, _ := json.Marshal(results)

	return shim.Success(resultsJSON)
}



//Host Change Application to change the host institute
func (s *SmartContract) hostchangeApplication(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {


	return shim.Success()

}




//Drop Out Application to let the student drop out of the consotium anytime
func (s *SmartContract) dropoutApplication(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	
	return shim.Success()
		
}




/********************************************************************************************************/

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
