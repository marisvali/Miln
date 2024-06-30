<!DOCTYPE html>
<html>
<body>

<h1>Miln collection script</h1>

<?php
$servername = "172.232.206.74";
$username = "playfulp_temp";
$password = "comeonthough";
$dbname = "playfulp_miln";

function LogInfo($message) {
	// file_put_contents("./submit-playthrough.log", "INFO: " . $message . "\n", FILE_APPEND);
}

function LogError($message) {
	file_put_contents("./submit-playthrough.log", "ERROR: " . $message . "\n", FILE_APPEND);
    http_response_code(513);
	die();
}

LogInfo("Start.");
if ($_SERVER['REQUEST_METHOD'] == 'POST') {
    LogInfo("Attempt to connect to database.");
    $conn = new mysqli($servername, $username, $password, $dbname);
    if ($conn->connect_error) {
        LogInfo("Connection failed: " . $conn->connect_error);
    } else {
        LogInfo("Connection succeeded!");
        
        $id = $_POST['id'];
        LogInfo("We got id: " . $id);
        if (isset($_FILES['playthrough'])) {
            $file = $_FILES['playthrough'];
            LogInfo("Found file.");
            $fileName = $file['name'];
            LogInfo("We got file name: " . $fileName);
            $fileTmpPath = $file['tmp_name'];
            LogInfo("We got file tmp path: " . $fileTmpPath);
            $fileContent = mysqli_real_escape_string($conn, file_get_contents($fileTmpPath));
            LogInfo("Read the file contents!");
            
            $sql = "UPDATE test4 SET playthrough = '$fileContent' WHERE id = '$id'";
        } else {
            $sql = "INSERT INTO test4 (id) VALUES ('$id')";
        }
        
        try {
            $conn->query($sql);
            LogInfo("Data successfully inserted!");
        } catch(Exception $e) {
            LogError("Error inserting data: " . $e->getMessage());
        }

        $conn->close();
    }
}
LogInfo("End.");
?>

</body>
</html>