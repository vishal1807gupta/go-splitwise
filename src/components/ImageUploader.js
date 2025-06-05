import React, { useState, useCallback } from 'react';
import { useDropzone } from 'react-dropzone'; // provides drag-and-drop file upload functionality

const ImageUploader = ({ groupId, onUploadSuccess, onCancel }) => {
    const [file, setFile] = useState(null);
    const [preview, setPreview] = useState(null);
    const [uploading, setUploading] = useState(false);
    const [error, setError] = useState(null);

    const onDrop = useCallback((acceptedFiles) => {
        setError(null);
        
        if (acceptedFiles.length === 0) {
            return;
        }
        
        const selectedFile = acceptedFiles[0];
        
        // Check file size (limit to 5MB)
        if (selectedFile.size > 5 * 1024 * 1024) {
            setError('File is too large. Please select an image under 5MB.');
            return;
        }
        
        // Create preview
        const objectUrl = URL.createObjectURL(selectedFile); // request to browser to create a temporary URL for the file(browser's file location -> URL)  -> stores the file in browser memory
        setPreview(objectUrl);
        // console.log(selectedFile, objectUrl);
        setFile(selectedFile);
        
        // Clean up preview URL when component unmounts
        return () => URL.revokeObjectURL(objectUrl);
    }, []);

    const { getRootProps, getInputProps, isDragActive } = useDropzone({
        onDrop,    // callback function triggered when files are dropped
        accept: {
            'image/*': ['.jpeg', '.jpg', '.png', '.gif', '.webp']
        },
        maxFiles: 1
    });

    const handleSubmit = async (e) => {
        e.preventDefault();

        // console.log(getRootProps(), getInputProps(), isDragActive);

        // to enable dropzone functionality these properties are used:
        // getRootProps() -> gives information(properties) about the dropzone area element
        // getInputProps() -> gives information(properties) about the input element of dropzone
        
        if (!file) {
            setError('Please select an image first');
            return;
        }
        
        setUploading(true);
        setError(null);
        
        const formData = new FormData(); // data format used to send file objects for http requests
        formData.append('image', file);
        formData.append('groupId', groupId);
        
        try {
            // Replace with your actual API endpoint
            const response = await fetch(`${process.env.REACT_APP_BACKEND_URL}/api/memories/upload`, {
                method: 'POST',
                body: formData,
            });
            
            if (!response.ok) {
                throw new Error('Upload failed');
            }
            
            const result = await response.json();
            
            if (onUploadSuccess) {
                onUploadSuccess(result.memory);
            }
            
            // Reset form
            setFile(null);
            setPreview(null);
        } catch (err) {
            console.error('Error uploading image:', err);
            setError('Failed to upload image. Please try again.');
        } finally {
            setUploading(false);
        }
    };

    const handleRemoveImage = () => {
        if (preview) {
            URL.revokeObjectURL(preview);
        }
        setFile(null);
        setPreview(null);
    };

    return (
        <div className="bg-white rounded-lg border border-gray-200">
            <form onSubmit={handleSubmit}>
                {!preview ? (
                    <div 
                        {...getRootProps()} 
                        className={`border-2 border-dashed rounded-lg p-6 cursor-pointer text-center transition-colors duration-200 ${
                            isDragActive 
                                ? 'border-indigo-500 bg-indigo-50' 
                                : 'border-gray-300 hover:border-indigo-400 hover:bg-gray-50'
                        }`}
                    >
                        <input {...getInputProps()} />
                        <div className="flex flex-col items-center justify-center space-y-3">
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                            </svg>
                            {isDragActive ? (
                                <p className="text-indigo-600">Drop the image here...</p>
                            ) : (
                                <div>
                                    <p className="text-gray-600 mb-1">Drag and drop an image here, or click to select</p>
                                    <p className="text-xs text-gray-500">JPG, PNG, GIF or WEBP (max 5MB)</p>
                                </div>
                            )}
                        </div>
                    </div>
                ) : (
                    <div className="relative">
                        <img 
                            src={preview} 
                            alt="Preview" 
                            className="w-full h-48 object-cover rounded-t-lg" 
                        />
                        <button 
                            type="button"
                            onClick={handleRemoveImage}
                            className="absolute top-2 right-2 bg-red-600 text-white p-1 rounded-full hover:bg-red-700 transition-colors duration-200"
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                                <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                            </svg>
                        </button>
                    </div>
                )}
                
                <div className="p-4">
                    
                    {error && (
                        <div className="mb-4 text-red-500 text-sm">
                            {error}
                        </div>
                    )}
                    
                    <div className="flex space-x-3">
                        <button
                            type="button"
                            onClick={onCancel}
                            disabled={uploading}
                            className="flex-1 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-50 transition-colors duration-200"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            disabled={!file || uploading}
                            className={`flex-1 py-2 rounded-md text-white transition-colors duration-200 ${
                                !file || uploading
                                    ? 'bg-indigo-400 cursor-not-allowed'
                                    : 'bg-indigo-600 hover:bg-indigo-700'
                            }`}
                        >
                            {uploading ? 'Uploading...' : 'Upload Memory'}
                        </button>
                    </div>
                </div>
            </form>
        </div>
    );
};

export default ImageUploader;