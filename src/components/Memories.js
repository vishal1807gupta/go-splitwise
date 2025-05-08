import React, { useState, useEffect } from 'react';
import ImageUploader from './ImageUploader';

const Memories = ({ groupId, onMemoryAdded }) => {
    const [memories, setMemories] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [showUploader, setShowUploader] = useState(false);
    const [selectedImage, setSelectedImage] = useState(null);
    const [showDeleteConfirm, setShowDeleteConfirm] = useState(null);
    const [isDeleting, setIsDeleting] = useState(false);

    useEffect(() => {
        fetchMemories();
    }, [groupId]);

    const fetchMemories = async () => {
        setLoading(true);
        try {
            // Replace with your actual API endpoint
            const response = await fetch(`${process.env.REACT_APP_BACKEND_URL}/api/memories/${groupId}`);
            if (!response.ok) {
                throw new Error('Failed to fetch memories');
            }
            const data = await response.json();
            setMemories(data.memories);
        } catch (err) {
            console.error('Error fetching memories:', err);
            setError('Failed to load memories. Please try again later.');
        } finally {
            setLoading(false);
        }
    };

    const handleDeleteMemory = async (memoryId) => {
        setIsDeleting(true);
        try {
            const response = await fetch(`${process.env.REACT_APP_BACKEND_URL}/api/memories/${memoryId}`, {
                method: 'DELETE',
            });
            
            if (!response.ok) {
                throw new Error('Failed to delete memory');
            }
            
            // Remove the deleted memory from state
            setMemories(memories.filter(memory => memory.id !== memoryId));
            setShowDeleteConfirm(null);
        } catch (err) {
            console.error('Error deleting memory:', err);
            setError('Failed to delete memory. Please try again.');
        } finally {
            setIsDeleting(false);
        }
    };

    const handleDownloadImage = async (imageUrl, memoryId, e) => {
        // Stop propagation to prevent opening the image viewer
        e.stopPropagation();
        
        try {
            // Fetch the image as a blob
            const response = await fetch(imageUrl);
            const blob = await response.blob();
            
            // Create a blob URL for the image
            const blobUrl = URL.createObjectURL(blob);
            
            // Create a link element
            const link = document.createElement('a');
            link.href = blobUrl;
            link.download = `group-memory-${memoryId}.jpg`; // Set filename for download
            
            // Append to the document, click, and remove
            document.body.appendChild(link);
            link.click();
            
            // Clean up
            document.body.removeChild(link);
            URL.revokeObjectURL(blobUrl); // Release the blob URL
        } catch (error) {
            console.error('Error downloading image:', error);
            // Optionally show an error message to the user
        }
    };

    const handleUploadSuccess = (newMemory) => {
        setMemories((prev) => [newMemory, ...prev]);
        setShowUploader(false);
        if (onMemoryAdded) onMemoryAdded();
    };

    // Enhanced loading state with nicer animation
    if (loading) {
        return (
            <div className="flex flex-col items-center justify-center py-10">
                <div className="w-16 h-16 border-t-4 border-indigo-500 border-solid rounded-full animate-spin"></div>
                <p className="mt-4 text-gray-600">Loading memories...</p>
            </div>
        );
    }

    if (error) {
        return (
            <div className="p-6 text-center bg-red-50 rounded-lg">
                <svg xmlns="http://www.w3.org/2000/svg" className="h-12 w-12 mx-auto text-red-500 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <p className="text-red-700 font-medium">{error}</p>
                <button 
                    onClick={fetchMemories}
                    className="mt-4 px-5 py-2 bg-indigo-600 hover:bg-indigo-700 text-white font-medium rounded-md shadow-sm transition-colors duration-200"
                >
                    Try Again
                </button>
            </div>
        );
    }

    // Function to close image viewer
    const closeImageViewer = () => {
        setSelectedImage(null);
    };

    return (
        <div className="flex flex-col h-full">
            {/* Main content */}
            <div className="flex-grow overflow-y-auto p-4 bg-white">
                {(!memories || memories.length === 0) ? (
                    <div className="text-center py-12">
                        <div className="mb-6">
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-16 w-16 mx-auto text-indigo-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                            </svg>
                        </div>
                        <h4 className="text-xl font-medium text-gray-700 mb-2">No memories yet</h4>
                        <p className="text-gray-500 mb-6 max-w-md mx-auto">Capture special moments with your group by adding photos to your shared memories collection.</p>
                        <button 
                            onClick={() => setShowUploader(true)}
                            className="px-6 py-3 bg-indigo-600 hover:bg-indigo-700 text-white font-medium rounded-lg shadow-md transition-colors duration-200 hover:shadow-lg"
                        >
                            Upload First Memory
                        </button>
                    </div>
                ) : (
                    <>
                        {/* Photos grid */}
                        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
                        {memories && memories.length > 0 && memories.map((memory) => (
                                <div 
                                    key={memory.id} 
                                    className="group relative rounded-xl overflow-hidden bg-white shadow-md hover:shadow-lg transition-all duration-300 aspect-square cursor-pointer"
                                    onClick={() => setSelectedImage(memory)}
                                >
                                    <img 
                                        src={memory.imageUrl} 
                                        alt="Group memory"
                                        className="w-full h-full object-cover transition-transform duration-500 group-hover:scale-110"
                                    />
                                    <div className="absolute inset-0 bg-black opacity-0 group-hover:opacity-20 transition-opacity duration-300"></div>
                                    
                                    {/* Action buttons container - positioned at top right */}
                                    <div className="absolute top-2 right-2 flex space-x-2 opacity-0 group-hover:opacity-90 transition-opacity duration-300 z-10">
                                        {/* Download button */}
                                        <button 
                                            className="bg-blue-600 text-white p-2 rounded-full hover:bg-blue-700"
                                            onClick={(e) => handleDownloadImage(memory.imageUrl, memory.id, e)}
                                        >
                                            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                                                <path fillRule="evenodd" d="M3 17a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm3.293-7.707a1 1 0 011.414 0L9 10.586V3a1 1 0 112 0v7.586l1.293-1.293a1 1 0 111.414 1.414l-3 3a1 1 0 01-1.414 0l-3-3a1 1 0 010-1.414z" clipRule="evenodd" />
                                            </svg>
                                        </button>
                                        
                                        {/* Delete button */}
                                        <button 
                                            className="bg-red-600 text-white p-2 rounded-full hover:bg-red-700"
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                setShowDeleteConfirm(memory.id);
                                            }}
                                        >
                                            <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                                                <path fillRule="evenodd" d="M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z" clipRule="evenodd" />
                                            </svg>
                                        </button>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </>
                )}
            </div>
            
            {/* Fixed footer with Add Photo button */}
            <div className="border-t bg-white p-4 flex justify-end">
                <button 
                    onClick={() => setShowUploader(true)}
                    className="flex items-center px-5 py-2.5 bg-indigo-600 hover:bg-indigo-700 text-white font-medium rounded-full shadow-md transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 mr-1.5" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z" clipRule="evenodd" />
                    </svg>
                    Add Photo
                </button>
            </div>
            
           {/* Image Viewer Modal */}
            {selectedImage && (
                <div 
                    id="image-viewer-modal" 
                    className="fixed inset-0 z-[200]"
                    onMouseDown={(e) => e.stopPropagation()}
                    onTouchStart={(e) => e.stopPropagation()}
                >
                    {/* Black overlay background */}
                    <div 
                        className="absolute inset-0 bg-black opacity-90"
                        onClick={closeImageViewer}
                    ></div>
                    
                    {/* Close button - fixed position */}
                    <button 
                        onClick={(e) => {
                            e.stopPropagation();
                            closeImageViewer();
                        }} 
                        className="fixed top-3 right-3 z-[201] bg-black bg-opacity-50 hover:bg-opacity-70 text-white w-6 h-6 rounded-full flex items-center justify-center"
                    >
                        <svg 
                            xmlns="http://www.w3.org/2000/svg" 
                            className="h-4 w-4" 
                            fill="none" 
                            viewBox="0 0 24 24" 
                            stroke="currentColor" 
                        >
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                    
                    {/* Image container with hidden scrollbars */}
                    <div 
                        className="fixed inset-0 z-[200] flex items-center justify-center"
                        onClick={closeImageViewer}
                        style={{ 
                            overflow: 'auto',
                            msOverflowStyle: 'none',  /* IE and Edge */
                            scrollbarWidth: 'none'     /* Firefox */
                        }}
                    >
                        {/* CSS to hide scrollbars in WebKit browsers (Chrome, Safari) */}
                        <style>
                            {`
                            #image-viewer-modal > div::-webkit-scrollbar {
                                display: none;
                            }
                            `}
                        </style>
                        
                        <div 
                            className="flex items-center justify-center min-h-full min-w-full"
                            onClick={(e) => e.stopPropagation()}
                        >
                            <img
                                src={selectedImage.imageUrl}
                                alt="Group memory"
                                className="m-auto"
                                style={{ 
                                    maxHeight: `${memories.length>3?'80vh':'50vh'}`,
                                    maxWidth: '90vw',
                                    objectFit: 'contain'
                                }}
                                onClick={(e) => e.stopPropagation()}
                                draggable="false"
                            />
                        </div>
                    </div>
                </div>
            )}

            {showDeleteConfirm && (
                <div className="fixed inset-0 z-[400] flex items-center justify-center">
                    <div 
                        className="absolute inset-0 bg-black bg-opacity-70"
                        onClick={() => setShowDeleteConfirm(null)}
                    ></div>
                    <div className="bg-white rounded-lg shadow-xl z-10 w-full max-w-md mx-4 p-5 relative">
                        <h3 className="text-xl font-bold text-gray-800 mb-4">Delete Memory?</h3>
                        <p className="text-gray-600 mb-6">Are you sure you want to delete this memory? This action cannot be undone.</p>
                        
                        <div className="flex space-x-3">
                            <button 
                                onClick={() => setShowDeleteConfirm(null)}
                                className="flex-1 py-2 bg-gray-200 hover:bg-gray-300 rounded-md text-gray-800 transition-colors"
                                disabled={isDeleting}
                            >
                                Cancel
                            </button>
                            <button 
                                onClick={() => handleDeleteMemory(showDeleteConfirm)}
                                className="flex-1 py-2 bg-red-600 hover:bg-red-700 rounded-md text-white transition-colors"
                                disabled={isDeleting}
                            >
                                {isDeleting ? 'Deleting...' : 'Delete'}
                            </button>
                        </div>
                    </div>
                </div>
            )}
            
            {/* Upload Modal */}
            {showUploader && (
                <div className="fixed inset-0 z-40 bg-black bg-opacity-50 flex items-center justify-center p-4">
                    <div className="bg-white rounded-xl shadow-xl w-full max-w-xl overflow-hidden transform transition-all">
                        <div className="p-4 bg-indigo-50 border-b border-indigo-100 flex justify-between items-center">
                            <h3 className="text-lg font-medium text-indigo-900">Add New Memory</h3>
                            <button 
                                onClick={() => setShowUploader(false)}
                                className="text-gray-500 hover:text-gray-700 focus:outline-none"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                                    <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                                </svg>
                            </button>
                        </div>
                        <div className="p-4">
                            <ImageUploader 
                                groupId={groupId}
                                onUploadSuccess={handleUploadSuccess}
                                onCancel={() => setShowUploader(false)}
                            />
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
};

export default Memories;