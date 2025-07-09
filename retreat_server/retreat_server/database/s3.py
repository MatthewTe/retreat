import boto3
from pathlib import Path
from loguru import logger
from botocore.exceptions import ClientError

def upload_db_to_s3(db_path: str) -> bool:
    s3_client = boto3.client("s3")

    if Path(db_path).is_file():
        try:
            response = s3_client.upload_file(db_path, "arn:aws:s3:::retreat-articles", "feed_db.sqlite")
        except ClientError as e:
            logger.error(e)
            return False

    return True