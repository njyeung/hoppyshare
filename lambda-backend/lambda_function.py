import json
from router import route_action

def handler(event, context):
    try:
        res = route_action(event)
        return {
            "statusCode": res.get("status_code", 500),
            "headers": {
                'Access-Control-Allow-Origin': '*',
                'Access-Control-Allow-Headers': 'Content-Type,Authorization,X-Amz-Date,X-Api-Key,X-Amz-Security-Token',
                'Access-Control-Allow-Methods': 'GET,POST,PUT,DELETE,OPTIONS',
                'Content-Type': 'application/json'
            },
            "body": json.dumps(res.get("json", {}))
        }

    except Exception as e:
        return {
            "statusCode": 500,
            "headers": {
                'Access-Control-Allow-Origin': '*',
                'Access-Control-Allow-Headers': 'Content-Type,Authorization,X-Amz-Date,X-Api-Key,X-Amz-Security-Token',
                'Access-Control-Allow-Methods': 'GET,POST,PUT,DELETE,OPTIONS',
                'Content-Type': 'application/json'
            },
            "body": json.dumps({"error": f"Internal error: {str(e)}"})
        }

if __name__ == "__main__":
    with open("event.json") as f:
        event = json.load(f)
    result = handler(event, None)
    print(json.dumps(result, indent=2))
