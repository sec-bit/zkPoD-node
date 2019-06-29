package main

const REQUEST_URL_LOCAL_HOST = "http://localhost"

const REQUEST_METHOD_POST = "POST"
const REQUEST_METHOD_GET = "GET"

const DEFAULT_BASIC_CONFGI_FILE = "./basic.json"
const DEFAULT_CONFIG_FILE = "./config.json"
const DEFAULT_KEYSTORE_FILE = "./keystore"
const DEFAULT_KEY_PATH = "./key"
const DEFAULT_PUBLISH_BIN_FILE = "./bin/pod_publish"
const DEFAULT_BOB_DIR = "./B"
const DEFAULT_ALICE_DIR = "./A"
const DEFAULT_NET_IP = "localhost:8888"
const DEFAULT_SERVER_PORT = "7777"
const DEFAULT_PODEX_CONTRACT_ADDRESS = "0x9785351A7B5b58fEd191c8B306b56358CaF2C065"
const DEFAULT_DB_PATH = "./pod.db"

const OPERATION_START = "start"
const OPERATION_ALICE_INITDATA = "initdata"
const OPERATION_ALICE_PUBLISH = "publish"
const OPERATION_ALICE_CLOSE = "close"
const OPERATION_WITHDRAW = "withdraw"
const OPERATION_BOB_PURCHASE = "purchase"
const OPERATION_BOB_DEPOSIT = "deposit"
const OPERATION_BOB_UNDEPOSIT = "undeposit"

const TRANSACTION_STATUS_START = "start"
const TRANSACTION_STATUS_START_FAILED = "startFailed"
const TRANSACTION_STATUS_NEGO = "nego"
const TRANSACTION_STATUS_NEGO_FAILED = "NegoFailed"
const TRANSACTION_STATUS_RECEIVED_REQUEST = "requested"
const TRANSACTION_STATUS_INVALID_REQUEST = "invalidRequest"
const TRANSACTION_STATUS_RECEIVED_REQUEST_FAILED = "requestFailed"
const TRANSACTION_STATUS_RECEIVED_RESPONSE = "responsed"
const TRANSACTION_STATUS_GENERATE_RESPONSE = "generateResponse"
const TRANSACTION_STATUS_SEND_RESPONSE_FAILED = "responseFailed"
const TRANSACTION_STATUS_RECEIVED_RECEIPT_FAILED = "receiptFailed"
const TRANSACTION_STATUS_RECEIPT = "receipted"
const TRANSACTION_STATUS_GENERATE_SECRET = "generateSecret"
const TRANSACTION_STATUS_GENERATE_SECRET_FAILED = "generateSecretFailed"
const TRANSACTION_STATUS_GOT_SECRET = "GotSecret"
const TRANSACTION_STATUS_SEND_SECRET_FAILED = "secretFailed"
const TRANSACTION_STATUS_GOT_SECRET_FAILED = "secretGotFailed"
const TRANSACTION_STATUS_SEND_SECRET_TERMINATED = "secretterminated"
const TRANSACTION_STATUS_SEND_CLIAM = "claimed"
const TRANSACTION_STATUS_SEND_CLIAM_FAILED = "claimFailed"
const TRANSACTION_STATUS_VERIFY_FAILED = "verifyFailed"
const TRANSACTION_STATUS_DECRYPT_FAILED = "decryptFailed"
const TRANSACTION_STATUS_CLOSED = "closed"
const TRANSACTION_STATUS_ERROR = "error"

const TRANSACTION_MODE_PLAIN_POD = "plain"
const TRANSACTION_MODE_TABLE_POD = "table"

const TRANSACTION_SUB_MODE_COMPLAINT = "complaint"
const TRANSACTION_SUB_MODE_ATOMIC_SWAP = "atomic_swap"
const TRANSACTION_SUB_MODE_ATOMIC_SWAP_VC = "atomic_swap_vc"
const TRANSACTION_SUB_MODE_VRF = "vrf"

const LOG_RESULT_SUCCESS = 1
const LOG_RESULT_FAILED = 2

const LOG_OPERATION_TYPE_CONFIG_SETTING = 1
const LOG_OPERATION_TYPE_ALICE_PUBLISH_INIT = 11
const LOG_OPERATION_TYPE_ALICE_PUBLISH = 12
const LOG_OPERATION_TYPE_ALICE_CLOSE = 13
const LOG_OPERATION_TYPE_ALICE_WITHDRAW_FROM_DATA = 14
const LOG_OPERATION_TYPE_ALICE_WITHDRAW_FROM_TX = 15
const LOG_OPERATION_TYPE_BOB_DEPOSIT = 21
const LOG_OPERATION_TYPE_BOB_UNDEPOSIT = 22
const LOG_OPERATION_TYPE_BOB_WITHDRAW = 23
const LOG_OPERATION_TYPE_BOB_TX = 24

// const LOG_OPERATION_TYPE_ALICE_TX_NEW = 31
// const LOG_OPERATION_TYPE_ALICE_TX_STEP_NEW = 32
// const LOG_OPERATION_TYPE_ALICE_TX_STEP_NEGO = 33
// const LOG_OPERATION_TYPE_ALICE_TX_STEP_REQUEST = 34
// const LOG_OPERATION_TYPE_ALICE_TX_STEP_RESPONSE = 35
// const LOG_OPERATION_TYPE_ALICE_TX_STEP_RECEIPT = 36
// const LOG_OPERATION_TYPE_ALICE_TX_STEP_SUBMIT_SECRET = 37
// const LOG_OPERATION_TYPE_BOB_TX_STEP_NEW = 41
// const LOG_OPERATION_TYPE_BOB_TX_STEP_NEGO = 42
// const LOG_OPERATION_TYPE_BOB_TX_STEP_REQUEST = 43
// const LOG_OPERATION_TYPE_BOB_TX_STEP_RESPONSE = 44
// const LOG_OPERATION_TYPE_BOB_TX_STEP_RECEIPT = 45
// const LOG_OPERATION_TYPE_BOB_TX_STEP_READ_SECRET = 46
// const LOG_OPERATION_TYPE_BOB_TX_STEP_DECRYPT = 47

const REQUEST_TIMEOUT = 60 * 5

const RESPONSE_SUCCESS = `{"code":"0","message":"%s"}`
const RESPONSE_FAILED_TO_RESPONSE = `{"code":"10001","message":"failed to read response"}`
const RESPONSE_DATA_NOT_EXIST = `{"code":"20001","message":"the data does not exist"}`
const RESPONSE_INITIALIZE_FAILED = `{"code":"20002","message":"failed to initialize config"}`
const RESPONSE_SAVE_CONFIG_FILE_FAILED = `{"code":"20003","message":"failed to save config file"}`
const RESPONSE_UPLOAD_KEY_FAILED = `{"code":"20004","message":"failed to upload key"}`
const RESPONSE_INCOMPLETE_PARAM = `{"code":"20005","message":"parameters is incomplete or invalid"}`
const RESPONSE_GENERATE_PUBLISH_FAILED = `{"code":"20006","message":"failed to generate publish data"}`
const RESPONSE_READ_DATABASE_FAILED = `{"code":"20008","message":"fail to read to db"}`
const RESPONSE_PUBLISH_TO_CONTRACT_FAILED = `{"code":"20010","message":"failed to publish to contract"}`
const RESPONSE_UNPUBLISH_TO_CONTRACT_FAILED = `{"code":"20011","message":"failed to unpublish to contract"}`
const RESPONSE_READ_CONTRACT_FAILED = `{"code":"20012","message":"failed to read contract"}`
const RESPONSE_DEPOSIT_CONTRACT_FAILED = `{"code":"20013","message":"failed to deposit eth from contract"}`
const RESPONSE_UNDEPOSIT_CONTRACT_FAILED = `{"code":"20014","message":"failed to undeposit eth from contract"}`
const RESPONSE_WITHDRAW_CONTRACT_FAILED = `{"code":"20015","message":"failed to withdraw eth from contract"}`
const RESPONSE_PURCHASE_FAILED = `{"code":"20016","message":"failed to purchase data"}`
const RESPONSE_TRANSACTION_FAILED = `{"code":"20017","message":"%s"}`
const RESPONSE_HAS_STARTED = `{"code":"20101","message":"the node has been started"}`

const RESPONSE_SAVE_FILE_FAILED = `{"code":"20026","message":"failed to save upload file"}`
const RESPONSE_NO_NEED_TO_WITHDRAW = `{"code":"20031","message":"no need to withdraw"}`

const RECOVERY_ERROR = `{"code":"10006","message":"internal error, please contact with the adminstrator"}`
